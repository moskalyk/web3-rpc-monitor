import * as dotenv from "dotenv";
import { createServer } from "http";
import { Server } from "socket.io";
import { ethers } from 'ethers'
import { fetch } from 'cross-fetch'

dotenv.config();

const httpServer = createServer();

const io = new Server(httpServer, {
    cors: {
      origin: "http://localhost:3001",
      methods: ["GET", "POST"]
    }
  });

io.on("connection", (socket) => {
  console.log('connected')
});

const provider_urls = [
    process!.env!.SEQUENCE_RPC!, 
    process!.env!.ALCHEMY_RPC!, 
    process!.env!.INFURA_RPC!, 
    process!.env!.QUICKNODE_RPC!, 
    process!.env!.POLYGON_RPC!, 
    process!.env!.ANKR_RPC!
]

const providers: any = []
let chains: any = []

provider_urls.map((url: string) => {
    providers.push(new ethers.providers.JsonRpcProvider(url))
})

setInterval(async () => {
    const blocks: any = []
    try {
        for(let i = 0; i < providers.length; i++) blocks.push(await providers[i].getBlockNumber())
        const res = await fetch('https://polygon-indexer.sequence.app/status')
        const json = (await res.json())
        blocks.push(json.checks.lastBlockNum)
        io.sockets.emit("data", {
            date: Date.now(), 
            blocks: blocks,
            max: Math.max(...blocks)
        });
        chains.push(blocks)
    }catch(err){
        console.log(err)
    }
}, 2000)

const getRatios = () => {
    let ratios: any = []
    let prior = Math.max(...chains[0])
    console.log('prior: ', prior)

    let last = Math.max(...chains[chains.length - 1])
    console.log('last: ', last)

    provider_urls.map((_: any, i: any) => {
        let count = 0
        chains.map((blocks: any) => {
            const maxHeight = Math.max(...blocks)
            count += Math.abs(blocks[i]-maxHeight)
        })
        ratios.push( count )
    })

    return [ratios, chains.length]
}
setInterval(async () => {
    const blocks: any = []
    try {
        chains = chains.slice(-(1*60*60/2))
        console.log(1*60*60/2)
        const [ratios, length] = getRatios()

        io.sockets.emit("live", {
            ratios: ratios,
            blocks: length
        });
    }catch(err){
        console.log(err)
    }
}, 5000)

httpServer.listen(5000);