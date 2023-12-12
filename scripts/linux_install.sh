#!/usr/bin/env bash
#check if we have sudo perms
if [[ `whoami` != "root" ]]; then
    echo "Script must be run as root for correct installation permissions"
    exit 1
fi

#grab the remote installer zip
curl https://endless-waltz-xyz-downloads.s3.us-east-2.amazonaws.com/ew_messenger_nix.zip -o /tmp/ew_messenger.zip

#unzip it to /tmp
cd /tmp && unzip ew_messenger

#copy our files into the correct locations
mv ew_messenger /usr/bin/ew_messenger
mv Icon.png /usr/share/ew.png
for i in `ls /home`; do 
    if [[ -d /home/$i/Desktop ]]; then
	dest=/home/$i/Desktop
	cp shortcuts/endlesswaltz.desktop $dest
	chmod a+x $dest/endlesswaltz.desktop
	if [[ `which gio` ]]; then
	    sudo -u $i -g $i dbus-launch gio set $dest/endlesswaltz.desktop "metadata::trusted" true
        fi
    fi
done

#modify file permissions
chmod +x /usr/bin/ew_messenger
chmod a+x /usr/share/ew.png
