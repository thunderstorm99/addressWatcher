# addressWatcher

## How to build
You can build this program using the following commandline
```
CGO_ENABLED=0 go build -ldflags '-s -w' -trimpath .
```

## How to use
Fill all needed parameters and execute the binary
```
./addressWatcher \
-address YOUR_ADDRESS \
-apikey YOUR_APIKEY_FOR_chainz.cryptoid.info \
-coin SHORT_FOR_YOUR_COIN \
-chatid TELEGRAM_CHATID \
-token TELEGRAM_BOT_TOKEN \
-name NICKNAME_OF_ADDRESS
```

## Prerequisits
1. Check if your Blockchain is available at https://chainz.cryptoid.info/
1. If so, take note of the short handle in the url for your coin (e.g. `btc` for bitcoin, as taken from `https://btc.cryptoid.info/btc/`)
1. You need to create a Telegram bot first by talking to `@botfather` in Telegram and following the instructions
1. Create a seperate group for you (,some others) and your bot in Telegram
1. Invite `@RawDataBot` into your channel and read out the chatid (including the `-` sign at the front) e.g. `-123456`
1. Remove `@RawDataBot` from your channel
1. Request an API key from https://chainz.cryptoid.info/api.key.dws
1. Start this program somewhere it can run forever, e.g. a server

## Use cases

### Monitoring a Masternode
This program can be used to track rewards for a masternode. If you put in the address of the masternode this program will alert you of any changes made to the wallet's amount

### Monitoring a donation address
If you have an address where people can donate cryptocurrency to, you can leave this program running to be alerted of any incoming donations


## Known issues

### API goes down
From time to time the API does not respond, or goes into maintenance mode. If this happens, the API responds with an amount of 0. My program will pick this up and think of it as a withdrawal of all funds.
I need to find a way to check the API's readyness first.