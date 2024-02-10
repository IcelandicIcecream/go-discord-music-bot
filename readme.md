Made a super simple bot to play youtube videos in Discord. A lot of bugs to fix, it's still a WIP. Here's a rundown on how to use it:

Assumming you have your discord and youtube api tokens set up and ready to go, add them to a .env file in the same directory. Refer to .env.example for the variable names.

To run the bot, type ```go run ./cmd/bot/main.go```

Commands:
```
"!play <youtube link>" - Downloads the mp3 from a youtube link and stores it in the tmp/ folder, then adds it to queue or plays it.
"!skip" - Skips the current song in queue
"!stop" - Stops the whole queue
```
