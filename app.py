import os
import time

from slackclient import SlackClient

token = os.environ['TOKEN']

sc = SlackClient(token)

if sc.rtm_connect():
	while True:
		for thing in sc.rtm_read():
			print thing
		time.sleep(0.5)
else:
	print 'Connection failed.'
