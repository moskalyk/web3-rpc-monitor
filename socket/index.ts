import * as dotenv from "dotenv";
import { createServer } from "http";
import { Server } from "socket.io";
import { ethers } from 'ethers'
import { fetch } from 'cross-fetch'

dotenv.config();

const httpServer = createServer();

const io = new Server(httpServer, {
    cors: {
      origin: "*",
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
        chains.push({blocks: blocks, time: new Date().toLocaleTimeString() } )
        console.log(`polling... #${Math.max(...blocks)}`)
    }catch(err){
        console.log(err)
    }
}, 2000)

const getRatios = () => {
    let ratios: any = []
    provider_urls.map((_: any, i: any) => {
        let count = 0
        chains.map((chain: any) => {
            const maxHeight = Math.max(...chain.blocks)
            count += Math.abs(chain.blocks[i]-maxHeight)
        })
        ratios.push( count )
    })
    return [ratios]
}

setInterval(async () => {
    const blocks: any = []
    try {
        chains = chains.slice(-(1*60*60/2))
        const [ratios] = getRatios()
        io.sockets.emit("live", {
            ratios: ratios
        });
    }catch(err){
        console.log(err)
    }
}, 2000)

setInterval(async () => {
    const blocks: any = []
    try {
        let day: any = {
            0: [],
            1: [],
            2: [],
            3: [],
            4: [],
            5: [],
        }

        let time: any = []
        
        chains.map((chain: any, i: any) => {
            day[0].push(chain.blocks[0])
            day[1].push(chain.blocks[1])
            day[2].push(chain.blocks[2])
            day[3].push(chain.blocks[3])
            day[4].push(chain.blocks[4])
            day[5].push(chain.blocks[5])
            time.push(chain.time)
        })

        io.sockets.emit("day", {
            blocks: day,
            time: time
        });
        
    }catch(err){
        console.log(err)
    }
}, 5000)

httpServer.listen(5000);