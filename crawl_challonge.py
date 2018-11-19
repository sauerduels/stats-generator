from requests import get
from bs4 import BeautifulSoup
from yaml import load, dump
import re

base_url = 'https://sauerduels.challonge.com/'
excluded = ['sd22', 'sd23']
links = []
output = {}

for i in range(1, 100):
    url = '{}?page={}'.format(base_url, i)
    print('Donwloading {}'.format(url))
    page = get(url).text
    soup = BeautifulSoup(page, 'html.parser')

    tournaments = soup.find_all('tr')
    if len(tournaments) == 0:
        break
        
    for tournament in tournaments:
        progress_bar = tournament.find('div', 'progress-bar')
        if progress_bar and '100' not in progress_bar.get('style'):
            continue
        for link in tournament.find_all('a'):
            if base_url in link.get('href'):
                links.append(link.get('href'))

for link in links:
    sd_pos = len(base_url)
    sd_name = link[sd_pos:sd_pos+4]
    if sd_name in excluded:
        continue
    output[sd_name] = []
    
    print('Donwloading {}'.format(link))
    page = get(link).text
    groups_html = re.findall(r'"scorecard_html":(["\'])((\\?.)*?)\1', page)
    
    for group_html in groups_html:
        participant_list = []
        soup = BeautifulSoup(group_html[1].encode('utf-8').decode('unicode_escape'), 'html.parser')
        
        advanced_labels = soup.find_all('span', {'class': 'label'})
        for label in advanced_labels:
            label.replaceWith('')
        
        participants = soup.find_all('td', 'participant')
        for participant in participants:
            participant_list.append(participant.text.strip().replace('\\xC3\\xB6', '\\xF6'))
        
        output[sd_name].append(participant_list)

with open('events.yml', 'w') as file:
    dump(output, file, default_flow_style=False)

