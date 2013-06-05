Deploying appfs
===============

Make the filesystem on the server.

	appfile -h <APP>.appspot.com -u dummy -p dummy mkfs


Replace the default password file.

	apppass <USERNAME> <PASSWORD> > password
	appfile -h <APP>.appspot.com -u dummy -p dummy write /.password < password
	rm password

Mount the appfs locally.

	appmount -h <APP>.appspot.com -u <USERNAME> -p <PASSWORD> $HOME/mnt
	umount $HOME/mnt
