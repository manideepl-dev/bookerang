import {test} from 'node:test';
import assert from 'node:assert/strict';
import {createApi} from './api.js';

test('all shared operations send correct operation type and return their field',async()=>{
 const original=globalThis.fetch;
 try {
   const operations={
    register:['mutation',{input:{email:'test@example.test'}}],login:['mutation',{email:'test@example.test',password:'example'}],
    me:['query',{}],myBooks:['query',{}],discover:['query',{latitude:12,longitude:77,radiusKm:5}],myRequests:['query',{}],
    messages:['query',{requestId:'r'}],addBook:['mutation',{input:{title:'Book'}}],
    archiveBook:['mutation',{bookId:'b'}],requestBook:['mutation',{bookId:'b'}],
    acceptRequest:['mutation',{requestId:'r'}],declineRequest:['mutation',{requestId:'r'}],
    cancelRequest:['mutation',{requestId:'r'}],markReturned:['mutation',{requestId:'r'}],
    sendMessage:['mutation',{requestId:'r',text:'Hello'}],blockUser:['mutation',{userId:'u'}],
    reportUser:['mutation',{requestId:'r',reason:'Concern'}]
   };
   const api=createApi('/graphql',()=> 'session');
   for(const [name,[kind,variables]] of Object.entries(operations)){
     globalThis.fetch=async(url,options)=>{
       assert.equal(url,'/graphql');assert.equal(options.method,'POST');
       assert.equal(options.headers.Authorization,'Bearer session');
       const body=JSON.parse(options.body);
       assert.ok(body.query.startsWith(kind));assert.ok(body.query.includes('{ '+name));
       assert.deepEqual(body.variables,variables);
       return {ok:true,json:async()=>({data:{[name]:name}})};
     };
     assert.equal(await api[name](variables),name);
   }
 } finally {globalThis.fetch=original;}
});
test('anonymous calls omit authorization and transport failures surface',async()=>{
 const original=globalThis.fetch;
 try{
   globalThis.fetch=async(url,options)=>{assert.equal(options.headers.Authorization,undefined);return {ok:false,status:503};};
   await assert.rejects(createApi('/graphql').me(),/HTTP 503/);
 }finally{globalThis.fetch=original;}
});
