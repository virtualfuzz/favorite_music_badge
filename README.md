<!-- SPDX-License-Identifier: AGPL-3.0-or-later -->

# favorite music badge

Preview at [https://codeberg.org/virtualfuzz](https://codeberg.org/virtualfuzz)

Takes your favorite music from your youtube music account and displays it into
your github/codeberg/gitlab readme!

## how does this work

We automatically scrape the youtube website to get the favorite music of a
specific user, this means the user has to set the channel's stats to public for
that to work. Then, we generate an image link from [shields.io](shields.io) and
we add it to the README.

## Installing

This is a go application, meaning it can be installed by running

`go install codeberg.org/virtualfuzz/favorite_music_badge@latest`

## Running

This will fetch the favorite music from a channel (please note the channel has
to show its stats publicly for that to work)\
`favorite_music_badge CHANNEL_ID`

This will automatically try to fetch the repository and will update the file
with the new favorite music obtained from the channel.\
`favorite_music_badge -repository "REPOSITORY_URL" -filename "README.md" CHANNEL_ID`

Please note that when updating, we need to find a
"FAVORITE_MUSIC_BADGE_AFTER_THIS_LINE", this tells where the favorite music
[<img src="https://img.shields.io/badge/Favorite%20music-Fils%20de%20joie%20by%20Stromae-mistyrose?style=for-the-badge&logo=youtube-music&labelColor=darkred&"/>](https://youtube.com/watch?v=M7Z2tgJo8Hg)
overwritten with the music badge.

## CICD/automatically updating

Because this project automatically scrapes the youtube music website, youtube
music isn't very happy and CICD will usually fail to scrape the website.

Because of that it is recommended to instead get your own server, and
automatically run the favorite_music_badge command every day, please note that
it downloads a repository inside of ./repository_to_modify so you will have to
remove it before re-running the command again.

favorite_music_badge runs without user input if a SSH key is set and is valid,
and if git is setup properly (username and email set).

## if you want to try out the gitlab cicd/github cicd anyway...

This will not work however and it will fail to scrape the youtube website.

[.gitlab-ci.yml](.gitlab-ci.yml) is the file for the gitlab cicd, simple change
the variables inside with the proper ones and you are good to go. Also add an
SSH_KEY which is a base64 encoded string of a private ssh key as a secret
variable. And a GIT_BOT_EMAIL which is an email as a secret variable as well.

It is the same for [.github/workflows/update.yml](.github/workflows/update.yml),
however this one wasn't tested and doesn't even seem to run.

## License

favorite_music_badge is licensed under the [AGPL-3.0-or-later](LICENSE.md).

Copyright (C) 2025 @virtualfuzz

This program is free software: you can redistribute it and/or modify it under
the terms of the GNU Affero General Public License as published by the Free
Software Foundation, either version 3 of the License, or (at your option) any
later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY
WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A
PARTICULAR PURPOSE. See the GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License along
with this program. If not, see <https://www.gnu.org/licenses/>.
