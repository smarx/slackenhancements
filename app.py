import os
import re
import time

from slackclient import SlackClient

token = os.environ['TOKEN']

sc = SlackClient(token)

user_id = sc.api_call('auth.test')['user_id']

stuff_to_do = []
def do_stuff():
	global stuff_to_do
	print stuff_to_do
	for thing in stuff_to_do:
		thing['count'] -= 1
		if thing['action'] == 'blink':
			sc.api_call('chat.update',
				token=token,
				channel=thing['channel'],
				ts=thing['timestamp'],
				text=thing['count'] % 2 == 0 and thing['text'] or ' ')
	stuff_to_do = filter(lambda thing: thing['count'] > 0, stuff_to_do)

if sc.rtm_connect():
	while True:
		for event in sc.rtm_read():
			if event['type'] == 'message' and event.get('user') == user_id:
				channel = event['channel']
				timestamp = event['ts']
				blink_match = re.match(r'^&lt;blink&gt;(.*)&lt;/blink&gt;$', event['text'])
				if blink_match:
					text = blink_match.groups()[0]
					stuff_to_do.append({
						'action': 'blink',
						'channel': channel,
						'timestamp': timestamp,
						'text': text,
						'count': 40,
					})
		time.sleep(0.25)
		do_stuff()
else:
	print 'Connection failed.'
