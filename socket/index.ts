import * as dotenv from "dotenv";
import { createServer } from "http";
import { Server } from "socket.io";
import { ethers } from 'ethers'
import { fetch } from 'cross-fetch'
import express from 'express'
import bodyParser from 'body-parser'
import cors from 'cors'
const Corestore = require("corestore")

dotenv.config();

const accountSid = process.env.TWILIO_ACCOUNT_SID;
const authToken = process.env.TWILIO_AUTH_TOKEN;
const client = require('twilio')(accountSid, authToken);

const PORT = process.env.PORT || 4000
const app = express();

const CLIENT_URL = 'http://localhost:3000'
const corsOptions = {
    origin: CLIENT_URL,
};
  
app.use(cors(corsOptions));
app.use(bodyParser.json())

const httpServer = createServer();
const corestore = new Corestore('./db')
const numbers = corestore.get({name: "numbers", valueEncoding: 'json'})
const behindOccurences = corestore.get({name: "behind_occurences", valueEncoding: 'json'});

(async () => {
    await behindOccurences.ready()
    if(behindOccurences.length == 0 || (await behindOccurences.get(behindOccurences.length - 1)).notifying == true){
        console.log(`initializing the log with :`);
        console.log((await behindOccurences.append({notifying: false})));
    }
})()
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
let isNotifying = false
let block_treshold = 20

provider_urls.map((url: string) => {
    providers.push(new ethers.providers.JsonRpcProvider(url))
})

const sendToNumbers = async (block_behind: number) => {
    const fullStream = numbers.createReadStream()
    for await (const number of fullStream) {
        client.messages
        .create({
           body: `Sequence Node Gateway blocks are behind by ${block_behind}`,
           from: '+16727020100',
           to: number.number
         })
        .then((message: any) => console.log(message.sid));
    }
}

setInterval(async () => {
    const blocks: any = []
    try {
        for(let i = 0; i < providers.length; i++) blocks.push(await providers[i].getBlockNumber())
        const res = await fetch('https://polygon-indexer.sequence.app/status')
        const json = (await res.json())
        blocks.push(json.checks.lastBlockNum)
        const max = Math.max(...blocks)

        if(!(await behindOccurences.get(behindOccurences.length - 1)).notifying) {
            console.log(Math.abs(blocks[0] - max))
            if(Math.abs(blocks[0] - max) >= block_treshold){
                await sendToNumbers(Math.abs(blocks[0] - max))
                setTimeout(async () => {
                    await behindOccurences.append({notifying: false})
                }, 10*60*1000)
                await behindOccurences.append({notifying: true})
            }
        }

        io.sockets.emit("data", {
            date: Date.now(), 
            blocks: blocks,
            max: max
        });
        chains.push({blocks: blocks, time: new Date().toLocaleTimeString() } )
        console.log(`polling... #${Math.max(...blocks)} ${(await behindOccurences.get(behindOccurences.length - 1)).notifying}`)
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

app.post('/signUp', async (req: any, res: any) => {
    console.log(req.body.number)
    res.send({
        status: 200,
        number: await numbers.append({number: req.body.number})
    })
})

app.listen(PORT, async () => {
    console.log(`listening on port: ${PORT}`)
})