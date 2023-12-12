#!/usr/bin/env bash
#check if we have sudo perms
if [[ `whoami` != "root" ]]; then
    echo "Script must be run as root for correct installation permissions"
    exit 1
fi

infile="/tmp/ew_messenger.zip"
if [[ -f $infile ]]; then
    echo "Removing stale install package..."
    rm $infile 
    rm /tmp/Icon.png
    rm /tmp/ew_messenger
    rm -rf /tmp/shortcuts
fi

#grab the remote installer zip
curl https://endless-waltz-xyz-downloads.s3.us-east-2.amazonaws.com/ew_messenger_nix.zip -o $infile

#unzip it to /tmp
cd /tmp && unzip ew_messenger.zip

if [[ -f /usr/bin/ew_messenger ]] || [[ -f /usr/share/ew.png ]]; then
	rm /usr/bin/ew_messenger
	rm /usr/share/ew.png
fi

#copy our files into the correct locations
mv ew_messenger /usr/bin/ew_messenger
mv Icon.png /usr/share/ew.png
for i in `ls /home`; do 
    dest=/home/$i/Desktop
    if [[ -d $dest ]] && [[ ! -f $dest/endlesswaltz.desktop ]]; then
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
