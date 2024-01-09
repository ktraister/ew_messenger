#!/usr/bin/env bash
#check if we have sudo perms
if [[ `whoami` != "root" ]]; then
    echo "Script must be run as root for correct installation permissions"
    exit 1
fi

cd /tmp
if [[ -f ew_messenger.zip ]]; then
    rm ew_messenger.zip
fi
if [[ -f ew_messenger ]]; then
    rm ew_messenger
fi
if [[ -f Icon.png ]]; then
    rm Icon.png
fi
if [[ -d shortcuts/ ]]; then
    rm -rf shortcuts
fi

#grab the remote installer zip
curl -s https://endless-waltz-xyz-downloads.s3.us-east-2.amazonaws.com/ew_messenger_nix.zip -o ew_messenger.zip
#unzip it to /tmp
unzip ew_messenger.zip

#copy our files into the correct locations
mv ew_messenger /usr/bin/ew_messenger
mv Icon.png /usr/share/ew.png
for i in `ls /home`; do 
    dest=/home/$i/Desktop
    if [[ -d $dest ]] && [[ ! -f $dest/endlesswaltz.desktop ]]; then
	cp shortcuts/endlesswaltz.desktop $dest
	chmod a+x $dest/endlesswaltz.desktop
	if [[ `which gio` ]] && [[ `which dbus-launch` ]]; then
	    sudo -u $i -g $i dbus-launch gio set $dest/endlesswaltz.desktop "metadata::trusted" true
        fi
    fi
done

#modify file permissions
chmod +x /usr/bin/ew_messenger
chmod a+x /usr/share/ew.png

echo
echo "Endless Waltz messenger has been installed at the latest available version!"
echo
