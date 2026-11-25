const USER = 'id displayName';
const BOOK = 'id title author description rentalPricePaise available distanceKm owner { '+USER+' }';
const REQUEST = 'id status note createdAt borrower { '+USER+' } book { '+BOOK+' }';
const MESSAGE = 'id text createdAt sender { '+USER+' }';

export function createApi(endpoint, getToken=()=>null) {
  async function execute(document, variables={}) {
    const token = getToken();
    const response = await fetch(endpoint, {
      method: 'POST',
      headers: {'Content-Type':'application/json', ...(token ? {Authorization:'Bearer '+token} : {})},
      body: JSON.stringify({query:document,variables})
    });
    if (!response.ok) throw new Error('HTTP '+response.status);
    const result = await response.json();
    if (result.errors?.length) throw new Error(result.errors[0].message);
    return result.data;
  }
  const operation = (name, defs, args, selection, mutation=false) => async (variables={}) => {
    const data=await execute((mutation?'mutation':'query')+(defs?'('+defs+')':'')+' { '+name+(args?'('+args+')':'')+(selection?' { '+selection+' }':'')+' }',variables);
    return data[name];
  };
  return {
    register:operation('register','$input:RegisterInput!','input:$input','token user { '+USER+' }',true),
    login:operation('login','$email:String!,$password:String!','email:$email,password:$password','token user { '+USER+' }',true),
    me:operation('me','','',USER),
    myBooks:operation('myBooks','','',BOOK),
    discover:operation('discover','$latitude:Float!,$longitude:Float!,$radiusKm:Float!,$search:String,$first:Int','latitude:$latitude,longitude:$longitude,radiusKm:$radiusKm,search:$search,first:$first',BOOK),
    myRequests:operation('myRequests','','',REQUEST),
    messages:operation('messages','$requestId:ID!,$first:Int,$before:ID','requestId:$requestId,first:$first,before:$before',MESSAGE),
    addBook:operation('addBook','$input:BookInput!','input:$input',BOOK,true),
    archiveBook:operation('archiveBook','$bookId:ID!','bookId:$bookId',BOOK,true),
    requestBook:operation('requestBook','$bookId:ID!,$note:String','bookId:$bookId,note:$note',REQUEST,true),
    acceptRequest:operation('acceptRequest','$requestId:ID!','requestId:$requestId',REQUEST,true),
    declineRequest:operation('declineRequest','$requestId:ID!','requestId:$requestId',REQUEST,true),
    cancelRequest:operation('cancelRequest','$requestId:ID!','requestId:$requestId',REQUEST,true),
    markReturned:operation('markReturned','$requestId:ID!','requestId:$requestId',REQUEST,true),
    sendMessage:operation('sendMessage','$requestId:ID!,$text:String!','requestId:$requestId,text:$text',MESSAGE,true),
    blockUser:operation('blockUser','$userId:ID!','userId:$userId','',true),
    reportUser:operation('reportUser','$requestId:ID!,$reason:String!,$details:String','requestId:$requestId,reason:$reason,details:$details','',true)
  };
}
