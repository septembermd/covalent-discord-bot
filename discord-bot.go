package main

import (
	"flag"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/x0xO/hhttp"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Token ODU5ODM1NjgzMTI4ODY4ODc1.YNyeYQ.gNb4f_Alu13lNGVqoeqi1XkSVl4
// https://discord.com/oauth2/authorize?client_id=859835683128868875&permissions=207936&scope=bot
var (
	Token string
	apiUrl = "https://api.coingecko.com/api/v3"
	lastPrice float64 = 0.001
	channel = "859836098913763362"
)

func init() {

	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func getCurrentPrice() float64 {
	client := hhttp.NewClient()

	headers := map[string]string{"Accept": "application/json"}
	request := client.Get(fmt.Sprintf("%s/coins/covalent", apiUrl)).SetHeaders(headers)
	response, err := request.Do()
	if err != nil {
		log.Panic(err)
	}

	var data Coins
	err = response.JSON(&data)
	if err != nil {
		log.Println(err)
	}

	return data.MarketData.CurrentPrice.Usd
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	// This isn't required in this specific example but it's a good practice.
	if m.Author.ID == s.State.User.ID {
		return
	}
	// If the message is "ping" reply with "Pong!"
	if m.Content == "ping" {
		s.ChannelMessageSend(m.ChannelID, "Pong!")
	}

	// If the message is "pong" reply with "Ping!"
	if m.Content == "pong" {
		s.ChannelMessageSend(m.ChannelID, "Ping!")
	}

	log.Println(m.ChannelID)
}

func PercentageChange(old float64, new float64) (delta float64) {
	diff := new - old
	delta = (diff / old) * 100
	return
}

func main() {

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// In this example, we only care about receiving message events.
	dg.Identify.Intents = discordgo.IntentsGuildMessages

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")

	c1 := make(chan string, 1)
	go func() {
		for {
			price := getCurrentPrice()
			delta := PercentageChange(lastPrice, price)
			if delta > 10 {
				c1 <- fmt.Sprintf("CQT is UP %v%%. Price is %v", delta, price)
			} else if delta < -10{
				c1 <- fmt.Sprintf("CQT is DOWN %v%%. Price is %v", delta, price)
			}
			lastPrice = price
			time.Sleep(10 * time.Minute)
		}
	}()

	select {
	case res := <-c1:
		_, err := dg.ChannelMessageSend(channel, res)
		if err != nil {
			log.Println(err)
		}
	}

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()
}