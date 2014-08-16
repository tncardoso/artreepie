package main

import (
	"code.google.com/p/go.net/context"
	"flag"
	"fmt"
	"github.com/garyburd/go-oauth/oauth"
	"github.com/xiam/twitter"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"html"
	"image"
	"image/png"
	"log"
	"math/rand"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	// Number of seconds to sleep after processing one timeline request.
	// This sleep executed even if images are processed.
	SECONDS_BETWEEN_CHECKS = 2 * 60
	// Maximum time of an image plot before a timeout happens. This is
	// used in order to avoid being stuck in an infinite loop.
	PLOT_TIMEOUT_MIN = 5
	// Name of the mongodb database that should be used for storing
	// information.
	DB = "artreepie"
	// Bot prefix that should be striped from the tweets.
	BOT_PREFIX = "@artreepie"
)

// Address of mongodb that should be used for storing generated art and
// processed tweets.
var mongodb = flag.String("mongodb", "localhost", "MongoDB address")

// When this flag is set to true artreepie runs in server mode. In
// server mode artreepie checks for mentions and replies with procedural
// art.
var server = flag.Bool("server", false, "Run artreepie in server mode")

// When plot flag is set to true, than artreepie expects three args one
// for each color R, G, B.
var plotf = flag.Bool("plot", false, "Use artreepie to plot image")

// Strip bot prefix from tweet. This function is used for removing the
// bot name from the tweet content leaving only 'code'. This function
// also cleans the content and unescape html entities.
func stripBotPrefix(tweet string) string {
	code := strings.Replace(tweet, BOT_PREFIX, "", 1)
	return html.UnescapeString(code)
}

// Tweets the response with the generated image. This function writes
// the image to disk before uploading it to twitter. Currently, the image
// name is hardcoded which means that you can generate one image at a time.
// In order to allow multiple images to be generated at once, they
// should be written to temporary file.
func tweetArt(ctx context.Context, screenName string, img *image.RGBA) error {
	twitter := ctx.Value("twitter").(*twitter.Client)

	imgFilename := "art-gen.png"
	imgFile, err := os.Create(imgFilename)
	defer imgFile.Close()
	if err != nil {
		return err
	}

	err = png.Encode(imgFile, img)
	if err != nil {
		return err
	}

	// send the actual update
	v := url.Values{}
	msg := messages[rand.Intn(len(messages))]
	_, err = twitter.UpdateWithMedia(
		fmt.Sprintf(msg, screenName),
		v, []string{imgFilename})
	return err
}

// Generate image using the r, g, b code snippets. If the image is
// successfully generated then it is tweeted to the author.
func generateArt(ctx context.Context, screenName, r, g, b string) {
	log.Printf("    Generating art:\n")
	log.Printf("        R: %s\n", r)
	log.Printf("        G: %s\n", g)
	log.Printf("        B: %s\n", b)

	sess := (ctx.Value("mongo").(*mgo.Session)).Clone()
	archiveCol := sess.DB(DB).C("Archive")

	// Create a new context with deadline. If the plot function takes
	// more than PLOTE_TIMEOUT_MIN minutes it is aborted with error.
	plotCtx, _ := context.WithDeadline(ctx,
		time.Now().Add(PLOT_TIMEOUT_MIN*time.Minute))

	img, err := plot(plotCtx, r, g, b)
	if err != nil {
		log.Printf("        Error generating user image: %s",
			err.Error())
	} else {
		// Store successful art in archive
		err = archiveCol.Insert(&Archive{
			User: screenName,
			R:    r,
			G:    g,
			B:    b,
		})
		if err != nil {
			log.Printf("        Error storing art: %s", err.Error())
		}

		// Tweet result to user
		err = tweetArt(ctx, screenName, img)
		if err != nil {
			log.Printf("        Error tweeting art: %s", err.Error())
		}
	}

}

// Check if this is an unseen post. If this code is new then it is saved
// in database. If this is the third tweet of a given author then the
// image is processed and uploaded.
func process(ctx context.Context, post map[string]interface{}) {
	sess := (ctx.Value("mongo").(*mgo.Session)).Clone()
	id := post["id_str"]
	screenName := post["user"].(map[string]interface{})["screen_name"].(string)
	text := post["text"].(string)

	log.Printf("Checking if post= %s screen_name= %s", id, screenName)

	// Try to fetch this id in Processed capped collection
	userCol := sess.DB(DB).C("User")
	processedCol := sess.DB(DB).C("Processed")
	n, err := processedCol.FindId(id).Count()
	if err != nil {
		log.Printf("    Error fetching count: %s\n", err.Error())
		return
	}

	if n > 0 {
		log.Printf("    Post already processed, skipping")
		return
	} else {
		// Mark this post as processed
		err = processedCol.Insert(bson.M{"_id": id})
		if err != nil {
			log.Printf("    Error storing. %s\n", err.Error())
			return
		}

		// Check if this is actually a code piece
		if !isCode(stripBotPrefix(text)) {
			log.Printf("    This is not code: %s", text)
			return
		}

		// Check other pieces of code of this author
		user := &User{}
		err := userCol.Find(bson.M{"_id": screenName}).One(&user)
		if err != nil {
			log.Printf("    Error reading user: %s\n", err.Error())
		}

		if err != nil || len(user.Data) < 2 {
			log.Printf("    New post. Storing in database")
			_, err := userCol.Upsert(bson.M{"_id": screenName},
				bson.M{"$push": bson.M{"data": bson.M{
					"text": text}}})
			if err != nil {
				log.Printf("    Error storing data. %s\n", err.Error())
				return
			}
		} else {
			// This user already have enough information for processing
			// Clear user data.
			err = userCol.UpdateId(screenName, bson.M{"$unset": bson.M{"data": 1}})
			if err != nil {
				log.Printf("    Error updating user: %s\n", err.Error())
				return
			}

			// Generate image and send to author
			generateArt(ctx, screenName,
				stripBotPrefix(user.Data[0].Text),
				stripBotPrefix(user.Data[1].Text),
				stripBotPrefix(text))
		}
	}
}

// Request the mentions timeline and check if there are new tweets to
// process.
func processMentions(ctx context.Context) {
	log.Printf("Requesting user mentions")

	twitter := (ctx.Value("twitter")).(*twitter.Client)
	v := url.Values{}
	data, err := twitter.MentionsTimeline(v)
	if err != nil {
		log.Printf("Error fetching mentions: %s\n", err.Error())
	}

	// iterate in reverse order to get older posts first
	for i := len(*data) - 1; i >= 0; i-- {
		// Check if mention should be processed
		tw := (*data)[i]
		process(ctx, tw)
	}
}

// Run bot. Continuously check for new mentions and accumulate code in
// order to plot images.
func run(ctx context.Context) {
	for {
		processMentions(ctx)
		time.Sleep(SECONDS_BETWEEN_CHECKS * time.Second)
	}
}

func main() {
	flag.Parse()

	log.Printf("Starting artreepie\n")
	ctx := context.Background()

	if *server {
		// Read credentials and create twitter client
		conf, err := readConfig()
		if err != nil {
			log.Printf("Error reading config: %s\n", err.Error())
			return
		}
		client := twitter.New(&oauth.Credentials{
			conf.App.Key, conf.App.Secret,
		})
		client.SetAuth(&oauth.Credentials{
			conf.User.Token, conf.User.Secret,
		})
		ctx = context.WithValue(ctx, "twitter", client)

		// Open mongo connection
		sess, err := mgo.Dial(*mongodb)
		if err != nil {
			log.Printf("Error connecting to mongo: %s\n", err.Error())
			return
		}
		ctx = context.WithValue(ctx, "mongo", sess)
		run(ctx)
	} else if *plotf {
		if len(flag.Args()) != 3 {
			log.Printf("You should specify one code snippet for each color")
		} else {
			img, err := plot(ctx, flag.Arg(0), flag.Arg(1), flag.Arg(2))
			if err != nil {
				log.Printf("Error creating image: %s", err.Error())
				return
			}
			imgFile, err := os.Create("result.png")
			defer imgFile.Close()
			if err != nil {
				log.Printf("Error creating file: %s", err.Error())
				return
			}

			err = png.Encode(imgFile, img)
			if err != nil {
				log.Printf("Error writing file: %s", err.Error())
				return
			}

			log.Printf("Done!")
		}
	} else {
		log.Printf("Please specify a mode: -server or -plot")
	}
}
