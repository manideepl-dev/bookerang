package dev.bookerang;

import org.springframework.stereotype.Controller;
import org.springframework.graphql.data.method.annotation.*;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.jdbc.core.JdbcTemplate;
import org.springframework.security.crypto.bcrypt.BCryptPasswordEncoder;
import org.springframework.transaction.annotation.Transactional;
import java.security.*;
import java.nio.charset.StandardCharsets;
import java.util.*;

@Controller
public class PilotController {
  private final JdbcTemplate db;
  private final BCryptPasswordEncoder passwords;
  private final String invite;
  private final SecureRandom random = new SecureRandom();
  public PilotController(JdbcTemplate db, BCryptPasswordEncoder passwords, @Value("${bookerang.invite}") String invite) {
    this.db=db; this.passwords=passwords; this.invite=invite;
  }
  private static IllegalArgumentException rejected() { return new IllegalArgumentException("Request not permitted or invalid"); }
  private static void require(boolean ok) { if(!ok) throw rejected(); }
  private static String clean(String value,int max) {
    require(value!=null && !value.isBlank() && value.length()<=max); return value.strip();
  }
  private static String optional(String value,int max) { if(value==null || value.isBlank()) return null; return clean(value,max); }
  private static UUID id(String value) { return UUID.fromString(value); }
  private static String hash(String value) {
    try { return HexFormat.of().formatHex(MessageDigest.getInstance("SHA-256").digest(value.getBytes(StandardCharsets.UTF_8))); }
    catch(NoSuchAlgorithmException e) { throw new IllegalStateException(e); }
  }
  private Map<String,Object> one(String sql,Object... args) {
    var rows=db.queryForList(sql,args); require(rows.size()==1); return rows.getFirst();
  }
  private UUID actor(String token) {
    require(token!=null && token.length()==43);
    return (UUID)one("SELECT account_id FROM session_token WHERE token_hash=? AND expires_at>now()",hash(token)).get("account_id");
  }
  // Deliberately serialize pilot writes across instances to make cross-entity checks atomic.
  // Replace this coarse lock with ordered entity-level locking before scaling.
  private void writeLock() { db.query("SELECT pg_advisory_xact_lock(761304)",rs -> {}); }
  private static Map<String,Object> obj(Object... pairs) {
    Map<String,Object> result=new LinkedHashMap<>(); for(int i=0;i<pairs.length;i+=2)result.put((String)pairs[i],pairs[i+1]); return result;
  }
  private Map<String,Object> user(UUID uid) {
    var row=one("SELECT id,display_name FROM account WHERE id=?",uid);
    return obj("id",uid.toString(),"displayName",row.get("display_name"));
  }
  private boolean blocked(UUID a,UUID b) {
    return Boolean.TRUE.equals(db.queryForObject("SELECT EXISTS(SELECT 1 FROM user_block WHERE (blocker_id=? AND blocked_id=?) OR (blocker_id=? AND blocked_id=?))",Boolean.class,a,b,b,a));
  }
  private Map<String,Object> session(UUID uid) {
    byte[] bytes=new byte[32]; random.nextBytes(bytes);
    String token=Base64.getUrlEncoder().withoutPadding().encodeToString(bytes);
    db.update("INSERT INTO session_token(token_hash,account_id,expires_at) VALUES(?,?,now()+interval '7 days')",hash(token),uid);
    return obj("token",token,"user",user(uid));
  }
  private static void coords(double lat,double lon) { require(Double.isFinite(lat)&&Double.isFinite(lon)&&lat>=-90&&lat<=90&&lon>=-180&&lon<=180); }
  private static int limit(Integer n,int fallback) { int value=n==null?fallback:n; require(value>=1&&value<=100);return value; }
  public record RegisterInput(String email,String password,String displayName,String inviteCode,double latitude,double longitude) {}
  public record BookInput(String title,String author,String description,int rentalPricePaise) {}
  @MutationMapping @Transactional
  public Map<String,Object> register(@Argument RegisterInput input) {
    require(invite.length()>=16 && input.inviteCode()!=null && MessageDigest.isEqual(invite.getBytes(StandardCharsets.UTF_8),input.inviteCode().getBytes(StandardCharsets.UTF_8)));
    coords(input.latitude(),input.longitude());
    String email=clean(input.email(),254).toLowerCase(Locale.ROOT);
    require(email.contains("@") && input.password()!=null && input.password().length()>=12 && input.password().getBytes(StandardCharsets.UTF_8).length<=72);
    UUID uid=UUID.randomUUID();
    db.update("INSERT INTO account(id,email,password_hash,display_name,location) VALUES(?,?,?,?,ST_SetSRID(ST_MakePoint(?,?),4326)::geography)",uid,email,passwords.encode(input.password()),clean(input.displayName(),80),Math.round(input.longitude()*100)/100.0,Math.round(input.latitude()*100)/100.0);
    return session(uid);
  }
  @MutationMapping @Transactional
  public Map<String,Object> login(@Argument String email,@Argument String password) {
    require(password!=null && password.getBytes(StandardCharsets.UTF_8).length<=72);
    var row=one("SELECT id,password_hash FROM account WHERE email=?",clean(email,254).toLowerCase(Locale.ROOT));
    require(passwords.matches(password,(String)row.get("password_hash")));
    return session((UUID)row.get("id"));
  }
  @QueryMapping public Map<String,Object> me(@ContextValue String token) { return user(actor(token)); }
  private static final String BOOKS="SELECT b.*, a.display_name, NOT b.archived AND NOT EXISTS(SELECT 1 FROM borrow_request r WHERE r.book_id=b.id AND r.status='ACCEPTED') AS available FROM book b JOIN account a ON a.id=b.owner_id ";
  private Map<String,Object> bookRow(Map<String,Object> row) {
    return obj("id",row.get("id").toString(),"title",row.get("title"),"author",row.get("author"),"description",row.get("description"),"rentalPricePaise",row.get("rental_price_paise"),"available",row.get("available"),"owner",obj("id",row.get("owner_id").toString(),"displayName",row.get("display_name")),"distanceKm",row.get("distance_km"));
  }
  private Map<String,Object> book(UUID bid) { return bookRow(one(BOOKS+" WHERE b.id=?",bid)); }
  @QueryMapping public List<Map<String,Object>> myBooks(@ContextValue String token) {
    return db.queryForList(BOOKS+" WHERE b.owner_id=? AND NOT b.archived ORDER BY b.title,b.id",actor(token)).stream().map(this::bookRow).toList();
  }
  @QueryMapping public List<Map<String,Object>> discover(@ContextValue String token,@Argument double latitude,@Argument double longitude,@Argument double radiusKm,@Argument String search,@Argument Integer first) {
    UUID uid=actor(token); coords(latitude,longitude);require(Double.isFinite(radiusKm)&&radiusKm>0&&radiusKm<=50);
    double lat=Math.round(latitude*100)/100.0, lon=Math.round(longitude*100)/100.0;
    String q=search==null?"":search.strip();require(q.length()<=150);
    String sql="SELECT b.*,a.display_name,true AS available, ROUND((ST_Distance(a.location,ST_SetSRID(ST_MakePoint(?,?),4326)::geography)/1000)::numeric,0)::double precision AS distance_km FROM book b JOIN account a ON a.id=b.owner_id WHERE b.owner_id<>? AND NOT b.archived AND NOT EXISTS(SELECT 1 FROM borrow_request r WHERE r.book_id=b.id AND r.status='ACCEPTED') AND NOT EXISTS(SELECT 1 FROM user_block x WHERE (x.blocker_id=? AND x.blocked_id=b.owner_id) OR (x.blocker_id=b.owner_id AND x.blocked_id=?)) AND ST_DWithin(a.location,ST_SetSRID(ST_MakePoint(?,?),4326)::geography,?) AND (strpos(lower(b.title),lower(?))>0 OR strpos(lower(b.author),lower(?))>0) ORDER BY distance_km,b.id LIMIT ?";
    return db.queryForList(sql,lon,lat,uid,uid,uid,lon,lat,radiusKm*1000,q,q,limit(first,30)).stream().map(this::bookRow).toList();
  }
  @MutationMapping @Transactional public Map<String,Object> addBook(@ContextValue String token,@Argument BookInput input) {
    UUID uid=actor(token),bid=UUID.randomUUID();require(input.rentalPricePaise()>=0);
    db.update("INSERT INTO book(id,owner_id,title,author,description,rental_price_paise) VALUES(?,?,?,?,?,?)",bid,uid,clean(input.title(),200),clean(input.author(),200),optional(input.description(),2000),input.rentalPricePaise());
    return book(bid);
  }
  @MutationMapping @Transactional public Map<String,Object> archiveBook(@ContextValue String token,@Argument String bookId) {
    UUID uid=actor(token),bid=id(bookId);writeLock();
    require(db.update("UPDATE book SET archived=true WHERE id=? AND owner_id=? AND NOT archived AND NOT EXISTS(SELECT 1 FROM borrow_request WHERE book_id=? AND status='ACCEPTED')",bid,uid,bid)==1);
    db.update("UPDATE borrow_request SET status='DECLINED' WHERE book_id=? AND status='PENDING'",bid);return book(bid);
  }
  private Map<String,Object> rawRequest(UUID rid) { return one("SELECT r.*,b.owner_id FROM borrow_request r JOIN book b ON b.id=r.book_id WHERE r.id=?",rid); }
  private boolean participant(Map<String,Object> row,UUID uid) { return uid.equals(row.get("owner_id"))||uid.equals(row.get("borrower_id")); }
  private UUID other(Map<String,Object> row,UUID uid) { require(participant(row,uid));return (UUID)(uid.equals(row.get("owner_id"))?row.get("borrower_id"):row.get("owner_id")); }
  private Map<String,Object> requestRow(Map<String,Object> row) {
    return obj("id",row.get("id").toString(),"book",book((UUID)row.get("book_id")),"borrower",user((UUID)row.get("borrower_id")),"status",row.get("status"),"note",row.get("note"),"createdAt",row.get("created_at").toString());
  }
  @QueryMapping public List<Map<String,Object>> myRequests(@ContextValue String token) {
    UUID uid=actor(token);
    return db.queryForList("SELECT r.*,b.owner_id FROM borrow_request r JOIN book b ON b.id=r.book_id WHERE r.borrower_id=? OR b.owner_id=? ORDER BY r.created_at DESC,r.id",uid,uid).stream().map(this::requestRow).toList();
  }
  @MutationMapping @Transactional public Map<String,Object> requestBook(@ContextValue String token,@Argument String bookId,@Argument String note) {
    UUID uid=actor(token),bid=id(bookId),rid=UUID.randomUUID();writeLock();
    var row=one("SELECT owner_id,archived FROM book WHERE id=?",bid);UUID owner=(UUID)row.get("owner_id");
    require(!uid.equals(owner)&&!blocked(uid,owner)&&Boolean.FALSE.equals(row.get("archived"))&&Boolean.TRUE.equals(book(bid).get("available")));
    db.update("INSERT INTO borrow_request(id,book_id,borrower_id,note,status) VALUES(?,?,?,?,'PENDING')",rid,bid,uid,optional(note,1000));return requestRow(rawRequest(rid));
  }
  private Map<String,Object> transition(String token,String requestId,String next) {
    UUID uid=actor(token),rid=id(requestId);writeLock();var row=rawRequest(rid);
    UUID owner=(UUID)row.get("owner_id"),borrower=(UUID)row.get("borrower_id"),bid=(UUID)row.get("book_id");
    String old=(String)row.get("status");
    switch(next) {
      case "ACCEPTED" -> { require(uid.equals(owner)&&old.equals("PENDING")&&!blocked(owner,borrower)&&Boolean.TRUE.equals(book(bid).get("available"))); }
      case "DECLINED" -> require(uid.equals(owner)&&old.equals("PENDING"));
      case "CANCELLED" -> require(uid.equals(borrower)&&old.equals("PENDING"));
      case "RETURNED" -> require(uid.equals(owner)&&old.equals("ACCEPTED"));
      default -> throw rejected();
    }
    require(db.update("UPDATE borrow_request SET status=? WHERE id=? AND status=?",next,rid,old)==1);
    if(next.equals("ACCEPTED")) db.update("UPDATE borrow_request SET status='DECLINED' WHERE book_id=? AND id<>? AND status='PENDING'",bid,rid);
    return requestRow(rawRequest(rid));
  }
  @MutationMapping @Transactional public Map<String,Object> acceptRequest(@ContextValue String token,@Argument String requestId) { return transition(token,requestId,"ACCEPTED"); }
  @MutationMapping @Transactional public Map<String,Object> declineRequest(@ContextValue String token,@Argument String requestId) { return transition(token,requestId,"DECLINED"); }
  @MutationMapping @Transactional public Map<String,Object> cancelRequest(@ContextValue String token,@Argument String requestId) { return transition(token,requestId,"CANCELLED"); }
  @MutationMapping @Transactional public Map<String,Object> markReturned(@ContextValue String token,@Argument String requestId) { return transition(token,requestId,"RETURNED"); }
  private Map<String,Object> messageRow(Map<String,Object> row) { return obj("id",row.get("id").toString(),"sender",user((UUID)row.get("sender_id")),"text",row.get("body"),"createdAt",row.get("created_at").toString()); }
  @QueryMapping public List<Map<String,Object>> messages(@ContextValue String token,@Argument String requestId,@Argument Integer first,@Argument String before) {
    UUID uid=actor(token),rid=id(requestId);var row=rawRequest(rid);
    require(participant(row,uid)&&List.of("ACCEPTED","RETURNED").contains(row.get("status")));
    int n=limit(first,50);
    List<Map<String,Object>> rows;
    if(before==null) rows=db.queryForList("SELECT * FROM message WHERE request_id=? ORDER BY created_at DESC,id DESC LIMIT ?",rid,n);
    else { UUID cursor=id(before);one("SELECT id FROM message WHERE id=? AND request_id=?",cursor,rid);
      rows=db.queryForList("SELECT * FROM message WHERE request_id=? AND (created_at,id)<(SELECT created_at,id FROM message WHERE id=? AND request_id=?) ORDER BY created_at DESC,id DESC LIMIT ?",rid,cursor,rid,n); }
    Collections.reverse(rows);return rows.stream().map(this::messageRow).toList();
  }
  @MutationMapping @Transactional public Map<String,Object> sendMessage(@ContextValue String token,@Argument String requestId,@Argument String text) {
    UUID uid=actor(token),rid=id(requestId),mid=UUID.randomUUID();writeLock();var row=rawRequest(rid);
    require(participant(row,uid)&&"ACCEPTED".equals(row.get("status"))&&!blocked(uid,other(row,uid)));
    db.update("INSERT INTO message(id,request_id,sender_id,body) VALUES(?,?,?,?)",mid,rid,uid,clean(text,2000));return messageRow(one("SELECT * FROM message WHERE id=?",mid));
  }
  @MutationMapping @Transactional public boolean blockUser(@ContextValue String token,@Argument String userId) {
    UUID uid=actor(token),target=id(userId);require(!uid.equals(target));writeLock();
    db.update("INSERT INTO user_block(blocker_id,blocked_id) VALUES(?,?) ON CONFLICT DO NOTHING",uid,target);
    db.update("UPDATE borrow_request r SET status='DECLINED' FROM book b WHERE r.book_id=b.id AND r.status='PENDING' AND ((b.owner_id=? AND r.borrower_id=?) OR (b.owner_id=? AND r.borrower_id=?))",uid,target,target,uid);
    return true;
  }
  @MutationMapping @Transactional public boolean reportUser(@ContextValue String token,@Argument String requestId,@Argument String reason,@Argument String details) {
    UUID uid=actor(token),rid=id(requestId);UUID target=other(rawRequest(rid),uid);
    db.update("INSERT INTO report(id,request_id,reporter_id,reported_id,reason,details) VALUES(?,?,?,?,?,?)",UUID.randomUUID(),rid,uid,target,clean(reason,100),optional(details,2000));return true;
  }
}
