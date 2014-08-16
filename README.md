# Artreepie

Artreepie is a procedural art bot for twitter. After tweeting three
consecutive code snippets to artreepie, an image is calculated and
tweeted.

You can check artreepie in action here:
[http://www.twitter.com/artreepie](http://www.twitter.com/artreepie)

## Installing

In order to install artreepie it is necessary to have the golang
environment. For more information, check:

[http://golang.org/doc/install](http://golang.org/doc/install)

With golang installing artreepie is as simple as:

    $ go install github.com/tncardoso/artreepie

If you want to use the server mode then 
[MongoDB](http://www.mongodb.org/downloads) is also needed.

## Configuring

In order to use artreepie in server mode it is necessary to have an app
registered in [Twitter](https://apps.twitter.com/). The app key and
secret can then be used for authenticating an user and getting his
credentials.

One way to obtain these credentials is using 
[xiam's](https://github.com/xiam/twitter) twitter client:

    $ go install github.com/xiam/twitter/cli/twitter
    $ twitter -key $APP_KEY -secret $APP_SECRET

A file named settings.json should be created with the app and user
credentials.

    {
        "app": 
        {
            "key": "$APP_KEY",
            "secret": "$APP_SECRET"
        },
        "user":
        {
            "token": "authenticated user token",
            "secret": "authenticated user secret"
        }
    }

## Running

Artreepie can be run in two modes:

#### Plot

This mode is used for calculating an image without running the full
server. Three arguments are needed, one code snippet for each color
channel.

    $ artreepy -plot "(+ i j)" "(& i j)" "(| i j)"

#### Server

In this mode artreepie checks for mentions and respond to code snippets 
with generated images.

    $ artreepy -server

## Contributions

There are different improvements that could be made to atreepie. If you
want to contribute consider doing one of these:

* Feedback to user on failed calculations
* Additional functions in twik scope
* Web gallery and status

## License

Artreepie is made available under the LGPLv3 license.
