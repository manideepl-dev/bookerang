import {test,expect} from '@playwright/test';
const user={id:'u1',displayName:'Reader'};
const book={id:'b1',title:'The Alchemist',author:'Paulo Coelho',description:null,rentalPricePaise:0,available:true,distanceKm:2,owner:{id:'u2',displayName:'Neighbor'}};
test('sign in, discover and request a nearby book',async({page,context})=>{
 await context.grantPermissions(['geolocation']);await context.setGeolocation({latitude:12.97,longitude:77.59});
 await page.route('**/graphql',async route=>{
   if(route.request().method()==='OPTIONS')return route.fulfill({status:204,headers:{'access-control-allow-origin':'*','access-control-allow-methods':'POST, OPTIONS','access-control-allow-headers':'authorization, content-type'}});
   const {query,variables}=route.request().postDataJSON();
   let data;
   if(query.includes(' login('))data={login:{token:'test-token',user}};
   else if(query.includes(' discover(')){expect(route.request().headers().authorization).toBe('Bearer test-token');expect(variables.radiusKm).toBe(5);data={discover:[book]};}
   else if(query.includes(' requestBook(')){expect(variables.bookId).toBe('b1');data={requestBook:{id:'r1',status:'PENDING'}};}
   else throw new Error('Unexpected operation '+query);
   await route.fulfill({json:{data},headers:{'access-control-allow-origin':'*'}});
 });
 await page.goto('/');
 await page.getByLabel('Email',{exact:true}).fill('reader@example.test');
 await page.getByLabel('Password',{exact:true}).fill('test-password');
 await page.getByRole('button',{name:'Sign in',exact:true}).click();
 await expect(page.getByText('Hello, Reader')).toBeVisible();
 await page.getByRole('button',{name:'Find books near me'}).click();
 await expect(page.getByText('The Alchemist')).toBeVisible();
 await page.getByRole('button',{name:'Request to borrow'}).click();
 await expect(page.getByText('Request sent. Check Requests for updates.')).toBeVisible();
});
test('invitation screen does not ask for location until submission',async({page})=>{
 await page.goto('/');
 await page.getByRole('button',{name:'Have an invitation? Join'}).click();
 await expect(page.getByLabel('Invitation code')).toBeVisible();
 await expect(page.getByRole('button',{name:'Join with current location'})).toBeVisible();
});
