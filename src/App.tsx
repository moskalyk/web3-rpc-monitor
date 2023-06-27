import React from "react";
import { Line } from "react-chartjs-2";
import "chartjs-plugin-streaming";
import moment from "moment";
import { io } from 'socket.io-client';
import { Bar } from 'react-chartjs-2';
import './App.css'

import { Box, Card, Text, Button, useTheme, SunIcon, IconButton} from '@0xsequence/design-system'

const Chart = require("react-chartjs-2").Chart;

const chartColors = {
  red: "rgb(255, 99, 132)",
  orange: "rgb(255, 159, 64)",
  yellow: "rgb(255, 205, 86)",
  green: "rgb(75, 192, 192)",
  blue: "rgb(54, 162, 235)",
  purple: "rgb(153, 102, 255)",
  grey: "rgb(201, 203, 207)"
};

let labels: any = []

const BlockCounts = () => {
  const socket = io('ws://216.128.182.90:5000');
  const [data, setData] = React.useState<any>([])
  const [blockCount, setBlockCount] = React.useState<any>(0)

  const options = {
    responsive: true,
    plugins: {
      legend: {
        position: 'top' as const,
      },
      title: {
        display: true,
        text: 'Chart.js Bar Chart',
      },
    },
  };


  React.useEffect(() => {

    socket.on('live', (packet: any) => {
      const fullLabels = [
        {
          label: 'Sequence',
          data: [],
          fill: false,
          borderColor: 'black',
          tension: 0.1,
        },
        {
          label: 'Alchemy',
          data: [],
          fill: false,
          borderColor: 'blue',
          tension: 0.1,
        },
        {
          label: 'Infura',
          data: [],
          fill: false,
          borderColor: 'orange',
          tension: 0.1,
        },
        {
          label: 'Quicknode',
          data: [],
          fill: false,
          borderColor: 'cyan',
          tension: 0.1,
        },
        {
          label: 'Polygon',
          data: [],
          fill: false,
          borderColor: 'purple',
          tension: 0.1,
        },
        {
          label: 'Ankr',
          data: [],
          fill: false,
          borderColor: 'lightblue',
          tension: 0.1,
        }
      ]

      const labels = fullLabels.map((label: any) => label.label)
      
      const data0 = {
        labels,
        datasets: [
          {
            label: 'Blocks',
            data: packet.ratios.sort((a: any, b: any) => a - b),
            backgroundColor: 'rgba(255, 99, 132, 0.5)',
          }
        ],
      };
      console.log(packet)
      setData(data0)
      setBlockCount(packet.blocks)
    })
  })

 

  return(
    <>
      <p style={{textAlign: 'center', width: '100%'}}>behind during # of blocks {blockCount}</p>{}
      <Bar options={options} data={data} />
    </>
  )
}

let count = 0
function LastDay (props: any) {
  const socket = io('ws://216.128.182.90:5000');
  const [init, setInit] = React.useState<any>(false)

  const [chartData, setChartData] = React.useState<any>({
    labels: [],
    datasets: [
      {
        label: 'Sequence',
        data: [],
        fill: false,
        borderColor: 'black',
        tension: 0.1,
      },
      {
        label: 'Alchemy',
        data: [],
        fill: false,
        borderColor: 'blue',
        tension: 0.1,
      },
      {
        label: 'Infura',
        data: [],
        fill: false,
        borderColor: 'orange',
        tension: 0.1,
      },
      {
        label: 'Quicknode',
        data: [],
        fill: false,
        borderColor: 'cyan',
        tension: 0.1,
      },
      {
        label: 'Polygon',
        data: [],
        fill: false,
        borderColor: 'purple',
        tension: 0.1,
      },
      {
        label: 'Ankr',
        data: [],
        fill: false,
        borderColor: 'lightblue',
        tension: 0.1,
      },
      {
        label: 'Sequence Indexer',
        data: [],
        fill: false,
        borderColor: 'pink',
        tension: 0.1,
      }
    ],
  });

  React.useEffect(() => {
        socket.on('day', (packet: any) => {
          console.log(packet.blocks[0])
      if(!init){
        console.log('tester ')

          // labels = [...(labels).slice(-1800), new Date().toLocaleTimeString()];
        // console.log(chartData)
        if(count <= 0){
          setChartData({
            labels: packet.time,
            datasets: [

              {
                ...chartData.datasets[0],
                data: [...chartData.datasets[0].data, ...packet.blocks[0]],
              },
              {
                ...chartData.datasets[1],
                data: [...chartData.datasets[1].data, ...packet.blocks[1]],
              },
              {
                ...chartData.datasets[2],
                data: [...chartData.datasets[2].data, ...packet.blocks[2]],
              },
              {
                ...chartData.datasets[3],
                data: [...chartData.datasets[3].data, ...packet.blocks[3]],
              },
              {
                ...chartData.datasets[4],
                data: [...chartData.datasets[4].data, ...packet.blocks[4]],
              },
              {
                ...chartData.datasets[5],
                data: [...chartData.datasets[5].data, ...packet.blocks[5]],
              }
            ],
          });
        }
        count++

        setInit(true)
      }

        })
    }, [init])
  return(
    <>
      <p>in the last #{chartData.datasets[0].data.length} blocks</p>
      <Line data={chartData} />
    </>
  )
}

function App() {
  const [init, setInit] = React.useState<any>(false)
  const [nav, setNav] = React.useState<any>(0)
  const socket = io('ws://216.128.182.90:5000');

  const [chartData, setChartData] = React.useState<any>({
    labels: [],
    datasets: [
      {
        label: 'Sequence',
        data: [],
        fill: false,
        borderColor: 'black',
        tension: 0.1,
      },
      {
        label: 'Alchemy',
        data: [],
        fill: false,
        borderColor: 'blue',
        tension: 0.1,
      },
      {
        label: 'Infura',
        data: [],
        fill: false,
        borderColor: 'orange',
        tension: 0.1,
      },
      {
        label: 'Quicknode',
        data: [],
        fill: false,
        borderColor: 'cyan',
        tension: 0.1,
      },
      {
        label: 'Polygon',
        data: [],
        fill: false,
        borderColor: 'purple',
        tension: 0.1,
      },
      {
        label: 'Ankr',
        data: [],
        fill: false,
        borderColor: 'lightblue',
        tension: 0.1,
      },
      {
        label: 'Sequence Indexer',
        data: [],
        fill: false,
        borderColor: 'pink',
        tension: 0.1,
      }
    ],
  });

  React.useEffect(() => {
    if(!init){

      socket.on('data', (packet: any) => {
        labels = [...(labels).slice(-10), new Date().toLocaleTimeString()];
        // console.log(chartData)
        setChartData({
          labels: labels,
          datasets: [

            {
              ...chartData.datasets[0],
              data: [...chartData.datasets[0].data.slice(-10), packet.blocks[0]-packet.max],
            },
            {
              ...chartData.datasets[1],
              data: [...chartData.datasets[1].data.slice(-10), packet.blocks[1]-packet.max],
            },
            {
              ...chartData.datasets[2],
              data: [...chartData.datasets[2].data.slice(-10), packet.blocks[2]-packet.max],
            },
            {
              ...chartData.datasets[3],
              data: [...chartData.datasets[3].data.slice(-10), packet.blocks[3]-packet.max],
            },
            {
              ...chartData.datasets[4],
              data: [...chartData.datasets[4].data.slice(-10), packet.blocks[4]-packet.max],
            },
            {
              ...chartData.datasets[5],
              data: [...chartData.datasets[5].data.slice(-10), packet.blocks[5]-packet.max],
            },
            {
              ...chartData.datasets[6],
              data: [...chartData.datasets[6].data.slice(-10), packet.blocks[6]-packet.max],
            },
          ],
        });
      });

      setInit(true)
    }
  }, [chartData]);

  const {theme, setTheme} = useTheme()

  const Compass = (nav: any) => {
    let navigator;
    console.log(nav)
    switch(nav){
      case 0:
        navigator = <Line data={chartData} />
        break;
      case 1:
        navigator = <BlockCounts />
        break;
      case 2:
        navigator = <LastDay/>
        break;
    }
    return navigator
  }

  React.useEffect(() => {

  }, [nav])

  return (
    <div className="App">
      <Box gap='6'>
        <IconButton style={{position: 'fixed', top: '20px', right: '20px'}} icon={SunIcon} onClick={() => {
          setTheme(theme == 'dark' ? 'light' : 'dark')
        }}/>
      </Box>
      <br/>
      <br/>
      <br/>
      <h1 className="title">polygon RPC monitor</h1>
      <p className="title" style={{textAlign: 'center', marginLeft: '-60px', cursor: 'pointer'}}><span onClick={() => {setInit(false);setNav(1)}}>block counts &nbsp;&nbsp;&nbsp;/</span><span onClick={() => {setInit(false);setNav(0)}}>&nbsp;&nbsp;&nbsp; live</span><span onClick={() => {setInit(false);setNav(2)}}>&nbsp;&nbsp;&nbsp; / &nbsp;&nbsp;&nbsp; day</span></p>
      <br/>
      {Compass(nav)}
    </div>
  );
}

export default App;