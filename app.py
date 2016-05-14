import os
import time

from slackclient import SlackClient

token = os.environ['TOKEN']

sc = SlackClient(token)

user_id = sc.api_call('auth.test')['user_id']

if sc.rtm_connect():
	while True:
		for event in sc.rtm_read():
			if event['type'] == 'message' and event['user'] == user_id:
				print 'I typed "%s" on channel "%s" with timestamp "%s"' % (
					event['text'],
					event['channel'],
					event['ts'])
		time.sleep(0.5)
else:
	print 'Connection failed.'
