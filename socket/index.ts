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
    }catch(err){
        console.log(err)
    }
}, 2000)

httpServer.listen(5000);