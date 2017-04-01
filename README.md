Feed Tube uses youtube-dl to download a full playlist or channel into a directory.  It can then generate a podcast RSS file which can be read by a podcatcher like [Podcast Addict](https://play.google.com/store/apps/details?id=com.bambuna.podcastaddict).

#### Requirements
youtube-dl with codecs necessary to encode to mp3.

#### Example
Feed Tube works best with a publicly accessible server.  If you control a server at podcast.awesomechannel.com running a webserver like Apache or Nginx with a doc root at /var/www, you could run this command:
```sh
feedTube channel  'AwesomeYoutubeChannel' \
--apiKey 'YOUR_YOUTUBE_API_KEY' \
--outputFolder '/var/www/podcasts/awesome' \
--baseURL 'https://podcast.awesomechannel.com/podcasts/awesome' \
--xmlFile '/var/www/podcasts/awesome.xml'
```

You could then add `https://podcast.awesomechannel.com/podcasts/awesome.xml` to your podcatcher and you can listen to your favorite YouTube channel.  You can even add that command to your crontab, and you'll automatically get new content as it is published.

#### API Key
For more information on getting a YouTube API key read [this](https://developers.google.com/youtube/v3/getting-started#before-you-start) or watch [this](https://youtu.be/Im69kzhpR3I).
