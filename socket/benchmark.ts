import fetch from 'cross-fetch';

let startTime: any, endTime: any;

const start = () => {
  startTime = new Date();
}

const end = () => {
  endTime = new Date();
  return Math.round(endTime - startTime)
};

(async () => {
    start()
    const res = await fetch('http://localhost:8000/api/rpc')
    console.log(end())
})()    