import React from 'react';
import {render,screen,fireEvent} from '@testing-library/react-native';
import Pilot from '../../shared/Pilot';
const endpoint='http://localhost/graphql';
const originalFetch=global.fetch;
afterEach(()=>{global.fetch=originalFetch;jest.restoreAllMocks();});
const reply=data=>Promise.resolve({ok:true,json:async()=>({data})});
test('sign in and request location only on discovery',async()=>{
 const locate=jest.fn().mockResolvedValue({latitude:12.97,longitude:77.59});
 global.fetch=jest.fn().mockImplementation((url,options)=>{
 const {query,variables}=JSON.parse(options.body);
 if(query.includes(' login('))return reply({login:{token:'token',user:{id:'u1',displayName:'Reader'}}});
 if(query.includes(' discover(')){expect(options.headers.Authorization).toBe('Bearer token');expect(variables.latitude).toBe(12.97);return reply({discover:[]});}
 throw Error('Unexpected operation');
 });
 render(<Pilot endpoint={endpoint} locate={locate}/>);
 fireEvent.changeText(screen.getByLabelText('Email'),'reader@example.test');
 fireEvent.changeText(screen.getByLabelText('Password'),'example-password');
 fireEvent.press(screen.getByText('Sign in'));
 await screen.findByText('Hello, Reader');expect(locate).not.toHaveBeenCalled();
 fireEvent.press(screen.getByText('Find books near me'));
 await screen.findByText('No matching books yet. Try another search or radius.');
 expect(locate).toHaveBeenCalledTimes(1);
});
test('denied location prevents registration request',async()=>{
 const locate=jest.fn().mockRejectedValue(new Error('Location denied'));
 global.fetch=jest.fn();
 render(<Pilot endpoint={endpoint} locate={locate}/>);
 fireEvent.press(screen.getByText('Have an invitation? Join'));
 fireEvent.press(screen.getByText('Join with current location'));
 await screen.findByText('Location denied');expect(global.fetch).not.toHaveBeenCalled();
});
test('authentication failure remains on sign-in screen',async()=>{
 global.fetch=jest.fn().mockResolvedValue({ok:true,json:async()=>({errors:[{message:'Denied'}]})});
 render(<Pilot endpoint={endpoint} locate={jest.fn()}/>);
 fireEvent.press(screen.getByText('Sign in'));
 await screen.findByText('Denied');
 expect(screen.queryByText('Find your next read')).toBeNull();
});
