import os
import re
import time

from slackclient import SlackClient

token = os.environ['TOKEN']

sc = SlackClient(token)

user_id = sc.api_call('auth.test')['user_id']

MAX_COUNT = 40

important = None

stuff_to_do = []
def do_stuff():
	global stuff_to_do
	print stuff_to_do
	for thing in stuff_to_do:
		thing['count'] -= 1
		text = thing['text']

		needs_update = False

		if 'marquee' in thing['actions']:
			text = thing['text'] + ' '*5
			how_many = (MAX_COUNT-thing['count']) % len(text)
			text = '`' + text[how_many:] + text[:how_many] + '`'
			if thing['count'] == 0:
				text = thing['text']
			needs_update = True
		if 'blink' in thing['actions'] and thing['count'] % 2 == 1:
			text = ' '
			needs_update = True

		if needs_update:
			sc.api_call('chat.update',
				token=token,
				channel=thing['channel'],
				ts=thing['ts'],
				text=text)

	stuff_to_do = filter(lambda thing: thing['count'] > 0, stuff_to_do)
	if important['count'] <= 0:
		important = None

if sc.rtm_connect():
	while True:
		for event in sc.rtm_read():
			if event['type'] == 'message' and event.get('user') == user_id:
				channel = event['channel']
				timestamp = event['ts']

				text = event['text']

				actions_found = []
				for action in ('blink', 'marquee', 'important'):
					new_text = re.sub(r'&lt;{0}&gt;(.*)&lt;/{0}&gt;'.format(action),
						r'\1', text)
					if text != new_text:
						actions_found.append(action)
						text = new_text

				thing = {
					'actions': actions_found,
					'channel': event['channel'],
					'ts': event['ts'],
					'text': match.groups()[0],
					'count': MAX_COUNT,
				}

				if actions_found:
					stuff_to_do.append(thing)
				if 'important' in actions_found:
					important = thing

		time.sleep(0.25)
		do_stuff()
else:
	print 'Connection failed.'
