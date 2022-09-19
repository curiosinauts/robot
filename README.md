# robot 
A Slack chatops bot written in golang.

# Design options
1. A chatops bot handles commands within itself.
2. A chatops bot delegates the command handling using exec.Command(...)

# Decision
I decided to go with latter because this simplifies robot design. I can focus on developing a CLI that can be bundled later during docker build phase of this bot.

# How it works
Let's say a user types `@robot platformctl add user foo@example.com` in Slack. This Slack chatops bot receives the message, parses it then it calls `exec.Command('platformctl', []string{'add', 'user', 'foo@example.com'})` 

# Keeping it simple
This slack chatops bot is just a message relayer. I can think of few enhancements but I would like to think of slack bot as just another way of interfacing with CLI. It's very similar to executing a command in terminal. This bot sticks to pure text messages for receiving commands.   

# Security consideration
This bot should be situated behind a reverse proxy like nginx with TLS cert. This bot does not serve TLS cert directly. Although that functinoality can be added but I chose not to. Your organization might require peer approval prior to allowing someone's command to run. That's slightly out of scope for this bot although it's not terribly difficult to add such functionality. I am sure someone out there already built a slack bot with that functionality. I just don't see an universally accepted user and approver process that fits everyone's usecase.

![](/assets/slack.png)