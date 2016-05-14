import os
import re
import time

from slackclient import SlackClient
from whelk import shell

token = os.environ['TOKEN']

sc = SlackClient(token)

user_id = sc.api_call('auth.test')['user_id']

MAX_COUNT = 40

important = None

stuff_to_do = []
def do_stuff():
	global important, stuff_to_do
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
	if important and important['count'] <= 0:
		important = None

if sc.rtm_connect():
	while True:
		for event in sc.rtm_read():
			if (important and
					event['type'] == 'message' and
					'user' in event and
					event['channel'] == important['channel'] and
					float(event['ts']) > float(important['ts'])):
				sc.api_call('chat.delete',
					token=token,
					channel=important['channel'],
					ts=important['ts'])
				important['ts'] = sc.api_call('chat.postMessage',
					token=token,
					channel=important['channel'],
					text=important['text'],
					as_user=True)['ts']
			if event['type'] == 'message' and event.get('user') == user_id:
				channel = event['channel']
				timestamp = event['ts']

				text = event['text']

				actions_found = []
				for action in ('blink', 'marquee', 'important', 'cow'):
					new_text = re.sub(r'&lt;{0}&gt;(.*)&lt;/{0}&gt;'.format(action),
						r'\1', text)
					if text != new_text:
						actions_found.append(action)
						text = new_text

				if 'cow' in actions_found:
					text = '```' + shell.cowsay(text)[1] + '```'

				thing = {
					'actions': actions_found,
					'channel': event['channel'],
					'ts': event['ts'],
					'text': text,
					'count': MAX_COUNT,
				}

				if 'cow' in actions_found:
					sc.api_call('chat.update',
						token=token,
						channel=thing['channel'],
						ts=thing['ts'],
						text=text)

				if actions_found:
					stuff_to_do.append(thing)
				if 'important' in actions_found:
					important = thing
					sc.api_call('chat.update',
						token=token,
						channel=thing['channel'],
						ts=thing['ts'],
						text=text)

		time.sleep(0.25)
		do_stuff()
else:
	print 'Connection failed.'
