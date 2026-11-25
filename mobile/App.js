import React from 'react';
import {Text,View} from 'react-native';
import {SafeAreaProvider,SafeAreaView} from 'react-native-safe-area-context';
import * as Location from 'expo-location';
import Pilot from '../shared/Pilot';
async function locate(){
 const p=await Location.requestForegroundPermissionsAsync();
 if(p.status!=='granted') throw new Error('Please allow location access and try again');
 const result=await Location.getCurrentPositionAsync({accuracy:Location.Accuracy.Balanced});
 return {latitude:result.coords.latitude,longitude:result.coords.longitude};
}
export default function App(){
 const endpoint=process.env.EXPO_PUBLIC_API_URL;
 return <SafeAreaProvider><SafeAreaView style={{flex:1,backgroundColor:'#f5f3ee'}}>{endpoint?<Pilot endpoint={endpoint} locate={locate}/>:<View style={{padding:24}}><Text>Set EXPO_PUBLIC_API_URL to your development server /graphql URL.</Text></View>}</SafeAreaView></SafeAreaProvider>;
}