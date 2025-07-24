#!/bin/sh
cd ~
eval "$(ssh-agent -s)"
ssh-add ~/.ssh/codeberg_profile_cicd
rm -rf ~/repository_to_modify
/home/jayden295/go/bin/favorite_music_badge -repository "ssh://git@codeberg.org/virtualfuzz/.profile.git" -filename README.md UCB3TKqt3XsvwZPeFw5skSjA
