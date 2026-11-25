import {test} from 'node:test';
import assert from 'node:assert/strict';
import {createApi} from './api.js';
test('discovery sends bearer and variables',async()=>{
  const original=globalThis.fetch;
  try {
    globalThis.fetch=async (url,options)=>{
      assert.equal(url,'http://localhost/graphql');
      assert.equal(options.headers.Authorization,'Bearer token');
      const body=JSON.parse(options.body);
      assert.match(body.query,/discover\(/);
      assert.equal(body.variables.radiusKm,5);
      return {ok:true,json:async()=>({data:{discover:[]}})};
    };
    const api=createApi('http://localhost/graphql',()=> 'token');
    assert.deepEqual(await api.discover({latitude:12,longitude:77,radiusKm:5}),[]);
  } finally {globalThis.fetch=original;}
});
test('GraphQL errors propagate',async()=>{
 const original=globalThis.fetch;
 try {
   globalThis.fetch=async()=>({ok:true,json:async()=>({errors:[{message:'Denied'}]})});
   await assert.rejects(createApi('/graphql').me(),/Denied/);
 } finally {globalThis.fetch=original;}
});
