package main

// User struct that stores intermediate user codes. Tweeted codes are
// accumulated here until the three tweets are seen (R, G, B). This also
// represents the structure of the documents stored in mongodb.
type User struct {
	// User screen name
	Id string `bson:"_id"`
	// List of submissions currently in queue
	Data []struct {
		// Textual content of the tweet
		Text string `bson:"text"`
	} `bson:"data"`
}

// Archive is the representation of a successful processed image. The
// code used to generate the image is stored along with the user that
// tweeted the codes.
type Archive struct {
	// User screen name
	User string `bson:"usr"`
	// Code used for Red
	R string `bson:"r"`
	// Code used for Green
	G string `bson:"g"`
	// Code used for Blue
	B string `bson:"b"`
}
