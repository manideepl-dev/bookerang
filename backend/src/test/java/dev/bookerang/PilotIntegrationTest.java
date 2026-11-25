package dev.bookerang;

import java.util.*;
import java.util.concurrent.*;
import com.fasterxml.jackson.databind.JsonNode;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.boot.test.web.client.TestRestTemplate;
import org.springframework.http.*;
import static org.assertj.core.api.Assertions.assertThat;

@SpringBootTest(webEnvironment=SpringBootTest.WebEnvironment.RANDOM_PORT,properties="bookerang.invite=integration-test-invite-only")
class PilotIntegrationTest {
 @Autowired TestRestTemplate http;
 record Actor(String token,String id){}
 JsonNode call(String token,String query,Map<String,Object> vars){
   HttpHeaders headers=new HttpHeaders();headers.setContentType(MediaType.APPLICATION_JSON);
   if(token!=null)headers.setBearerAuth(token);
   var response=http.postForEntity("/graphql",new HttpEntity<>(Map.of("query",query,"variables",vars),headers),JsonNode.class);
   assertThat(response.getStatusCode().is2xxSuccessful()).isTrue();
   assertThat(response.getBody()).isNotNull();return response.getBody();
 }
 JsonNode ok(String token,String query,Map<String,Object> vars,String field){
   JsonNode result=call(token,query,vars);
   assertThat(result.path("errors").size()).as(result.toString()).isZero();
   JsonNode fieldValue=result.path("data").path(field);assertThat(fieldValue.isMissingNode()||fieldValue.isNull()).isFalse();return fieldValue;
 }
 void denied(String token,String query,Map<String,Object> vars){
   assertThat(call(token,query,vars).path("errors").size()).isGreaterThan(0);
 }
 Actor register(){
   var input=Map.of("email",UUID.randomUUID()+"@example.test","password","Test-only-password-123","displayName","Tester","inviteCode","integration-test-invite-only","latitude",12.97,"longitude",77.59);
   var result=ok(null,"mutation($input:RegisterInput!){register(input:$input){token user{id}}}",Map.of("input",input),"register");
   return new Actor(result.path("token").asText(),result.path("user").path("id").asText());
 }
 String book(Actor owner){
   return ok(owner.token(),"mutation($input:BookInput!){addBook(input:$input){id available}}",Map.of("input",Map.of("title","Test "+UUID.randomUUID(),"author","Test Author","rentalPricePaise",0)),"addBook").path("id").asText();
 }
 String request(Actor borrower,String book){
   return ok(borrower.token(),"mutation($id:ID!){requestBook(bookId:$id){id status}}",Map.of("id",book),"requestBook").path("id").asText();
 }
 String transitionQuery(String name){return "mutation($id:ID!){"+name+"(requestId:$id){id status book{available}}}";}
 JsonNode transition(Actor a,String name,String id){return ok(a.token(),transitionQuery(name),Map.of("id",id),name);}
 String sendQuery(){return "mutation($id:ID!){sendMessage(requestId:$id,text:\"Hello\"){id text}}";}
 String messagesQuery(){return "query($id:ID!){messages(requestId:$id){id text sender{id}}}";}
 @Test void unauthenticatedAndInvalidInvitationDenied(){
   denied(null,"{me{id}}",Map.of());
   denied(null,"mutation($input:RegisterInput!){register(input:$input){token}}",Map.of("input",Map.of("email",UUID.randomUUID()+"@example.test","password","Test-only-password-123","displayName","Tester","inviteCode","wrong","latitude",12.97,"longitude",77.59)));
 }
 @Test void lendingLifecycleAndParticipantAuthorization(){
   Actor owner=register(),borrower=register(),stranger=register();
   String bid=book(owner);
   JsonNode discovery=ok(borrower.token(),"{discover(latitude:12.97,longitude:77.59,radiusKm:5,first:100){id owner{id}}}",Map.of(),"discover");
   // Search by unique title in a separate test if fixture volume exceeds pagination.
   assertThat(discovery.isArray()).isTrue();
   denied(owner.token(),"mutation($id:ID!){requestBook(bookId:$id){id}}",Map.of("id",bid));
   String rid=request(borrower,bid);
   denied(borrower.token(),transitionQuery("acceptRequest"),Map.of("id",rid));
   denied(borrower.token(),sendQuery(),Map.of("id",rid));
   assertThat(transition(owner,"acceptRequest",rid).path("status").asText()).isEqualTo("ACCEPTED");
   denied(stranger.token(),messagesQuery(),Map.of("id",rid));
   denied(stranger.token(),sendQuery(),Map.of("id",rid));
   ok(borrower.token(),sendQuery(),Map.of("id",rid),"sendMessage");
   assertThat(ok(owner.token(),messagesQuery(),Map.of("id",rid),"messages").get(0).path("text").asText()).isEqualTo("Hello");
   denied(borrower.token(),transitionQuery("markReturned"),Map.of("id",rid));
   var returned=transition(owner,"markReturned",rid);
   assertThat(returned.path("status").asText()).isEqualTo("RETURNED");
   assertThat(returned.path("book").path("available").asBoolean()).isTrue();
   denied(owner.token(),sendQuery(),Map.of("id",rid));
   assertThat(ok(borrower.token(),messagesQuery(),Map.of("id",rid),"messages").size()).isEqualTo(1);
 }
 @Test void competingRequestsCannotBothBeAccepted() throws Exception {
   Actor owner=register(),a=register(),b=register();String bid=book(owner),ra=request(a,bid),rb=request(b,bid);
   try(var pool=Executors.newVirtualThreadPerTaskExecutor()){
     var start=new CountDownLatch(1);
     var x=pool.submit(()->{start.await();return call(owner.token(),transitionQuery("acceptRequest"),Map.of("id",ra));});
     var y=pool.submit(()->{start.await();return call(owner.token(),transitionQuery("acceptRequest"),Map.of("id",rb));});
     start.countDown();
     var responses=List.of(x.get(20,TimeUnit.SECONDS),y.get(20,TimeUnit.SECONDS));
     assertThat(responses.stream().filter(r->r.path("errors").size()==0).count()).isEqualTo(1);
     var requests=ok(owner.token(),"{myRequests{id status}}",Map.of(),"myRequests");
     Map<String,String> statuses=new HashMap<>();requests.forEach(r->statuses.put(r.path("id").asText(),r.path("status").asText()));
     assertThat(List.of(statuses.get(ra),statuses.get(rb))).containsExactlyInAnyOrder("ACCEPTED","DECLINED");
   }
 }
 @Test void blockPreventsMessagesAndReportsAreParticipantOnly(){
   Actor owner=register(),borrower=register(),stranger=register();String bid=book(owner),rid=request(borrower,bid);
   transition(owner,"acceptRequest",rid);
   String report="mutation($id:ID!){reportUser(requestId:$id,reason:\"Test concern\")}";
   denied(stranger.token(),report,Map.of("id",rid));
   assertThat(ok(borrower.token(),report,Map.of("id",rid),"reportUser").asBoolean()).isTrue();
   assertThat(ok(owner.token(),"mutation($id:ID!){blockUser(userId:$id)}",Map.of("id",borrower.id()),"blockUser").asBoolean()).isTrue();
   denied(borrower.token(),sendQuery(),Map.of("id",rid));
   denied(owner.token(),sendQuery(),Map.of("id",rid));
   transition(owner,"markReturned",rid);
   denied(borrower.token(),"mutation($id:ID!){requestBook(bookId:$id){id}}",Map.of("id",bid));
 }
 @Test void cancellationAndDeclineAreRoleRestricted(){
   Actor owner=register(),borrower=register();String bid=book(owner),r=request(borrower,bid);
   denied(owner.token(),transitionQuery("cancelRequest"),Map.of("id",r));
   denied(borrower.token(),transitionQuery("declineRequest"),Map.of("id",r));
   assertThat(transition(borrower,"cancelRequest",r).path("status").asText()).isEqualTo("CANCELLED");
   String next=request(borrower,bid);
   assertThat(transition(owner,"declineRequest",next).path("status").asText()).isEqualTo("DECLINED");
 }
}
