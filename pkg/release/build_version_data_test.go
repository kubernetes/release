/*
Copyright 2020 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package release_test

const testJobCache = `
[
  {
    "buildnumber": "1259149965448450048",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259165313757351938",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259180666998755328",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259196016138129409",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259211368150601729",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259226718804119555",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259242071479291904",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259257421105205248",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259272775005114369",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259288124207403009",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259303474592485376",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259318825115979778",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259334176604164096",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259349529472274433",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259364878469042177",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259380230271799298",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259395581235695618",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259410933424328705",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259426283733913600",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259441634219659265",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259456986332794880",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259472337183444993",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259487688021512194",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259503040596021249",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259518390691696640",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259533744344141826",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259549095333203968",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259564443398836224",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259579794324983809",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259595146412953600",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259610497385238528",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259625850278514688",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259641199510163458",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259656552327942144",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259671903266672640",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259687253400096768",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259702603982311424",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259717955252391936",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259733305842995200",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259748659344445441",
    "job-version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "version": "v1.18.3-beta.0.41+ec280c2f9e131c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259764007972114433",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259779361754583041",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259794712613621760",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259810063661404160",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259825413354426370",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259840766570663936",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259856116204965888",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259871469614141440",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259886820489957376",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259902170073927681",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259917524468764672",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259932839810437121",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259948190023553024",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259963539850792962",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259978891955539972",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1259994243791851520",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260009594172739585",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260024944457158656",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260040296759037954",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260055646728884228",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260070997898301440",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260086349424234498",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260101701478649859",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260117051490439169",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260132403167367169",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260147756614291457",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260163104981913601",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260178457132797953",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260193807417217024",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260209158527913985",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260224509923823616",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260239862573830145",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260255212606590977",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260270563172028416",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260285916354711553",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260301266890788868",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260316618341224450",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260331969321897985",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260347319950249984",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260362672700919808",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260378022331027458",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260393373852766208",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260408724313346050",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260424075436625920",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260439427717533697",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260454778773704705",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260470130412883968",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260485480760217600",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260500832764301313",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260516182281162753",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260531535983939585",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260546885987340295",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260562237081260033",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260577588166791168",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260592939919216642",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260608290266550272",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260623642924945408",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260640006909726722",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260655350722334721",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260670701044502529",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260685897486045186",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260701246738665472",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260716597681590275",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260731949769560064",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260747334933811202",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260762652532019200",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260778002694803456",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260793353738391555",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260808705314656257",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260824055871705089",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260839407271809030",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260854758634164224",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260870109874884612",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260885460729729024",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260900811353886720",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260916163873869827",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260931514388975616",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260946866577608706",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260962216190939139",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260977567691706368",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1260992918898872323",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261008205702500354",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261023555940782080",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261038906363613184",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261054258170564611",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261069609021214720",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261084962199703553",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261100311158722560",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261115663401881603",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261131015154307074",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261146365333868544",
    "job-version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "version": "v1.18.3-beta.0.43+209a8b052eadd3",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261161716314542081",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261177067966304258",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261192418993115136",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261207769839570944",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261223122586046464",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261238472845299712",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261253823846944768",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261269174546599936",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261284532816973828",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261299877996924933",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261315228394590211",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261330578792255489",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261345931106717698",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261361281911230464",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261376633529438208",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261391984023572484",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261407335423676416",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261422685867479040",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261438037930283011",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261453390538346496",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261468739778383872",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261484091358842880",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261499442461151234",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261515169599590401",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261530395359318023",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261545746713284609",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261561097891090432",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261576448964038660",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261591800569663488",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261607150686310400",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261622504632356864",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261637853901754373",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261653204462997504",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261668555569500161",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261683907070267393",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261699257967054849",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261714609627205632",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261729960469467141",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261745310888103936",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261760662690861065",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261776014065799171",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261791364505407491",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261806715448332288",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261822067800543232",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261837418512781313",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261852769719947264",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261868120495099906",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261883471328972800",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261898824259997697",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261914174569582594",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261929524564594690",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261944876321214466",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261960227691958272",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261975578861375488",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1261990929997238277",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262006280340377600",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262021631639818240",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262036984210132995",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262052333693440000",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262067684980297728",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262083037185708033",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262098387101028353",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262113739507765251",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262129090496827393",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262144440730914818",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "FAILURE",
    "passed": false
  },
  {
    "buildnumber": "1262159791669645312",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262175143392710656",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262190494599876608",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "FAILURE",
    "passed": false
  },
  {
    "buildnumber": "1262205847296020481",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262221196817076225",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262236547797749761",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262251898526765061",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262267250455351297",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262282601826095105",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262297952194400257",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262313303472869376",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262328655279820801",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262344006398906368",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262359357618655232",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262375463444025347",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262390814214983691",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262406167288614912",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262421516675452934",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262436868205580290",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262452219014287362",
    "job-version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "version": "v1.18.3-beta.0.46+1666bcfa4d3518",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262467570334699520",
    "job-version": "v1.18.3-beta.0.48+8aa020340748c8",
    "version": "v1.18.3-beta.0.48+8aa020340748c8",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262482921667694594",
    "job-version": "v1.18.3-beta.0.48+8aa020340748c8",
    "version": "v1.18.3-beta.0.48+8aa020340748c8",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262498273759858689",
    "job-version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262513623582904324",
    "job-version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262528976111276032",
    "job-version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262544326500552707",
    "job-version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262559677556723722",
    "job-version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262575028684197888",
    "job-version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262590379958472708",
    "job-version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262605731023032327",
    "job-version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262621082339250180",
    "job-version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262636432963407881",
    "job-version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262651783856001029",
    "job-version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262667135566483456",
    "job-version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262682487696396296",
    "job-version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262697837808848902",
    "job-version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262713188881797121",
    "job-version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262728540667777024",
    "job-version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262743891023499268",
    "job-version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262759243497345025",
    "job-version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262774593538494468",
    "job-version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262789947077693447",
    "job-version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262805295915077639",
    "job-version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262820647273238528",
    "job-version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262835761795829764",
    "job-version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "result": "FAILURE",
    "passed": false
  },
  {
    "buildnumber": "1262851113049133057",
    "job-version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "result": "FAILURE",
    "passed": false
  },
  {
    "buildnumber": "1262866464315019268",
    "job-version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262881814926594052",
    "job-version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262897166272172032",
    "job-version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "version": "v1.18.3-beta.0.52+0ca4dc5a186ed0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262912516405596162",
    "job-version": "v1.18.3-beta.0.55+40774f46e6a6c2",
    "version": "v1.18.3-beta.0.55+40774f46e6a6c2",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262927869214986249",
    "job-version": "v1.18.3-beta.0.55+40774f46e6a6c2",
    "version": "v1.18.3-beta.0.55+40774f46e6a6c2",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262943218501160960",
    "job-version": "v1.18.3-beta.0.58+d6e40f410ca91c",
    "version": "v1.18.3-beta.0.58+d6e40f410ca91c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262958571004366851",
    "job-version": "v1.18.3-beta.0.58+d6e40f410ca91c",
    "version": "v1.18.3-beta.0.58+d6e40f410ca91c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262973921058099205",
    "job-version": "v1.18.3-beta.0.58+d6e40f410ca91c",
    "version": "v1.18.3-beta.0.58+d6e40f410ca91c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1262989272240099336",
    "job-version": "v1.18.3-beta.0.58+d6e40f410ca91c",
    "version": "v1.18.3-beta.0.58+d6e40f410ca91c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263004624093188100",
    "job-version": "v1.18.3-beta.0.58+d6e40f410ca91c",
    "version": "v1.18.3-beta.0.58+d6e40f410ca91c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263019974385995778",
    "job-version": "v1.18.3-beta.0.58+d6e40f410ca91c",
    "version": "v1.18.3-beta.0.58+d6e40f410ca91c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263035325517664258",
    "job-version": "v1.18.3-beta.0.58+d6e40f410ca91c",
    "version": "v1.18.3-beta.0.58+d6e40f410ca91c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263050677781794826",
    "job-version": "v1.18.3-beta.0.58+d6e40f410ca91c",
    "version": "v1.18.3-beta.0.58+d6e40f410ca91c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263066028401758209",
    "job-version": "v1.18.3-beta.0.58+d6e40f410ca91c",
    "version": "v1.18.3-beta.0.58+d6e40f410ca91c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263081379210465281",
    "job-version": "v1.18.3-beta.0.58+d6e40f410ca91c",
    "version": "v1.18.3-beta.0.58+d6e40f410ca91c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263096730283413504",
    "job-version": "v1.18.3-beta.0.58+d6e40f410ca91c",
    "version": "v1.18.3-beta.0.58+d6e40f410ca91c",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263112081427664896",
    "job-version": "v1.18.4-rc.0.1+d2bb709f9aa035",
    "version": "v1.18.4-rc.0.1+d2bb709f9aa035",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263127433326891009",
    "job-version": "v1.18.4-rc.0.1+d2bb709f9aa035",
    "version": "v1.18.4-rc.0.1+d2bb709f9aa035",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263142784668274688",
    "job-version": "v1.18.4-rc.0.1+d2bb709f9aa035",
    "version": "v1.18.4-rc.0.1+d2bb709f9aa035",
    "result": "FAILURE",
    "passed": false
  },
  {
    "buildnumber": "1263158135766388744",
    "job-version": "v1.18.4-rc.0.1+d2bb709f9aa035",
    "version": "v1.18.4-rc.0.1+d2bb709f9aa035",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263173486474432515",
    "job-version": "v1.18.4-rc.0.1+d2bb709f9aa035",
    "version": "v1.18.4-rc.0.1+d2bb709f9aa035",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263188838872780800",
    "job-version": "v1.18.4-rc.0.1+d2bb709f9aa035",
    "version": "v1.18.4-rc.0.1+d2bb709f9aa035",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263204189329166338",
    "job-version": "v1.18.4-rc.0.1+d2bb709f9aa035",
    "version": "v1.18.4-rc.0.1+d2bb709f9aa035",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263219540305645570",
    "job-version": "v1.18.4-rc.0.1+d2bb709f9aa035",
    "version": "v1.18.4-rc.0.1+d2bb709f9aa035",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263234890883665924",
    "job-version": "v1.18.4-rc.0.1+d2bb709f9aa035",
    "version": "v1.18.4-rc.0.1+d2bb709f9aa035",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263250242120192004",
    "job-version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263265593830674435",
    "job-version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263280944178008066",
    "job-version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263296296572162056",
    "job-version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263311646592339969",
    "job-version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263326997862420482",
    "job-version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263342349660983296",
    "job-version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263357700243197952",
    "job-version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263373051131596800",
    "job-version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263388402368122882",
    "job-version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263403753319436288",
    "job-version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263419105596149762",
    "job-version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263434455431778315",
    "job-version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263449806727024648",
    "job-version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263465159288950786",
    "job-version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263480509212659712",
    "job-version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263495861329989633",
    "job-version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263511213506039808",
    "job-version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263526549785677824",
    "job-version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263541899688415232",
    "job-version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263557251277262848",
    "job-version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263572601943363588",
    "job-version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263587953603514370",
    "job-version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "result": "FAILURE",
    "passed": false
  },
  {
    "buildnumber": "1263607582593912837",
    "job-version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263622933452951552",
    "job-version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263638285129879552",
    "job-version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263653636186050560",
    "job-version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263668987963641857",
    "job-version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263684338331947010",
    "job-version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263699690029846530",
    "job-version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263715041354452993",
    "job-version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263730391580151809",
    "job-version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "result": "SUCCESS",
    "passed": true
  },
  {
    "buildnumber": "1263745742963478529",
    "job-version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "version": "v1.18.4-rc.0.3+3ff09514d162b0",
    "result": "SUCCESS",
    "passed": true
  }
]
`
