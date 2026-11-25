'use client';
import React from 'react';
import Pilot from '../../shared/Pilot';
const locate=()=>new Promise((resolve,reject)=>{
 if(!navigator.geolocation) return reject(new Error('Location is unavailable'));
 navigator.geolocation.getCurrentPosition(p=>resolve({latitude:p.coords.latitude,longitude:p.coords.longitude}),()=>reject(new Error('Please allow location access and try again')),{timeout:15000,maximumAge:0});
});
export default function Page(){return <Pilot endpoint={process.env.NEXT_PUBLIC_API_URL||'http://localhost:8080/graphql'} locate={locate}/>;}