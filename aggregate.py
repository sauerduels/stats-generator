import pandas as pd
import glob
import os
import sys
from yaml import load, dump, SafeLoader
import json
import shutil
from glicko2 import Glicko2, WIN, DRAW, LOSS
import csv

banned_players = ['Rudi']
mode_names = {'0': 'ffa', '3': 'insta', '5': 'effic'}
groups = None

with open('events.yml', 'r') as file:
    groups = load(file, Loader=SafeLoader)

def validate_event(event_dataframe):
    # make sure that only one mode is played in one event
    unique_modes = event_dataframe['mode'].unique()
    if len(unique_modes) > 1:
        return 'More than one mode in event: {}'.format(', '.join(str(v) for v in list(unique_modes)))
    
    # make sure that no games are played on justice
    justice_games = event_dataframe.loc[event_dataframe['map'] == 'justice']
    if len(justice_games) > 0:
        return 'Justice game(s) found'
    
    # make sure that no more than 2 players have played in a single game
    game_groups = event_dataframe.groupby('timestamp')
    for group in game_groups.groups:
        group_players = game_groups.get_group(group)['player']
        if len(group_players) != 2:
            return 'Invalid number of players in {} [{}]'.format(group, ', '.join(list(group_players)))
    return None

def validate_player(player_name, player_set, group):
    # check that player has played against every other player in the group
    for player in group:
        if player != player_name and player not in player_set:
            return '{} not in {}\'s set ({})'.format(player, player_name, player_set)
    
    return None

def validate_groups(event_name, games):
    # check that we have group info for event
    if event_name not in groups:
        return 'Groups not found'
    
    # check that each player has played all their games
    for group in groups[event_name]:
        for player in group:
            validation = validate_player(player, games[player], group)
            if validation is not None:
                return validation
    
    return None

glicko2_env = Glicko2(tau=0.5)
stats = {'0': dict(), '3': dict(), '5': dict()}
ratings = {'0': dict(), '3': dict(), '5': dict()}

def add_stats(player_row, win):
    # create stats entry from game
    player_name = str(player_row['player'])
    game_mode = str(player_row['mode'])
    player_stats = player_row[['frags', 'deaths', 'damage', 'damage_dealt', 'suicides',
                        'damage_0', 'damage_dealt_0', 'damage_1', 'damage_dealt_1',
                        'damage_2', 'damage_dealt_2', 'damage_3', 'damage_dealt_3',
                        'damage_4', 'damage_dealt_4', 'damage_5', 'damage_dealt_5', 
                        'damage_6', 'damage_dealt_6']]
    player_stats['wins'] = 1 if win else 0
    player_stats['losses'] = 0 if win else 1
    player_stats['games'] = 1
    # account for special frag value for forfeits
    if player_stats['frags'] >= 1000:
        player_stats['frags'] = 0
    # add game to player's stats entries
    if player_name not in stats[game_mode]:
        stats[game_mode][player_name] = list()
    stats[game_mode][player_name].append(player_stats)
    # create player's glicko2 entry if not exists
    if player_name not in ratings[game_mode]:
        ratings[game_mode][player_name] = glicko2_env.create_rating(1500, 350, 0.06)

# get names of all events in logs directory
events_names = sorted([d for d in os.listdir('logs')])
for event_name in events_names:
    # read all log files for event
    event_files = glob.glob(os.path.join('logs', event_name, "*.csv"))
    event_dataframe = pd.concat((pd.read_csv(f,
                                             sep=' ',
                                             header=None,
                                             names=['timestamp', 'mode', 'map', 'player', 'frags',
                                                    'deaths', 'damage', 'damage_dealt', 'suicides', 
                                                    'damage_0', 'damage_dealt_0', 'damage_1', 'damage_dealt_1',
                                                    'damage_2', 'damage_dealt_2', 'damage_3', 'damage_dealt_3',
                                                    'damage_4', 'damage_dealt_4', 'damage_5', 'damage_dealt_5', 
                                                    'damage_6', 'damage_dealt_6'])
                                 for f in event_files), ignore_index=True)
    
    # validate log files and print any errors
    vaidation_error = validate_event(event_dataframe)
    if vaidation_error:
        print('Error in {}: {}'.format(event_name, vaidation_error))
        sys.exit(1)
    
    # group log entries by game (timestamp)
    game_groups = event_dataframe.groupby('timestamp')
    rating_series = dict()
    games = dict()
    mode = None
    
    # process each duel
    for group in game_groups.groups:
        # first stats entry
        g0 = game_groups.get_group(group).iloc[0]
        # second stats entry
        g1 = game_groups.get_group(group).iloc[1]
        # add stats to player's stats entries
        add_stats(g0, g0.frags > g1.frags)
        add_stats(g1, g0.frags <= g1.frags)
        
        winner_name = g0['player'] if g0.frags > g1.frags else g1['player']
        loser_name = g0['player'] if g0.frags <= g1.frags else g1['player']
        mode = str(g0['mode'])
        
        # add winner game entry
        if winner_name not in games:
            games[winner_name] = set()
        games[winner_name].add(loser_name)
        
        # add loser game entry
        if loser_name not in games:
            games[loser_name] = set()
        games[loser_name].add(winner_name)
        
        # add winner's entry to their glicko2 rating series
        if winner_name not in rating_series:
            rating_series[winner_name] = list()
        rating_series[winner_name].append((WIN, ratings[mode][loser_name]))
        
        # add loser's entry to their glicko2 rating series
        if loser_name not in rating_series:
            rating_series[loser_name] = list()
        rating_series[loser_name].append((LOSS, ratings[mode][winner_name]))
    
    # validate game groups. this guards against missing forfeits
    vaidation_error = validate_groups(event_name, games)
    if vaidation_error:
        print('Warning in {} groups: {}'.format(event_name, vaidation_error))
    
    # rate glicko2 for event
    for player in rating_series:
        ratings[mode][player] = glicko2_env.rate(ratings[mode][player], rating_series[player])

total_stats = dict()

# process total mode stats for each mode
for mode in stats:
    mode_stats = list()
    for player in stats[mode]:
        # ignore banned players
        if player in banned_players:
            continue
        # sum all stats entries for player
        player_stats = pd.DataFrame().append(stats[mode][player]).sum()
        # add player's name to series
        player_stats['player'] = player
        # add player's elo from glicko2 ratings
        player_stats['elo'] = ratings[mode][player].mu
        mode_stats.append(player_stats)
    # sort player stats by elo and name
    total_stats[mode] = pd.DataFrame(mode_stats).sort_values(['elo', 'player'], ascending=[False, True])
    # set ranks
    total_stats[mode]['rank'] = range(1, len(total_stats[mode]) + 1)
    # set index to player column so that we can add the tables together to calculate combined stats
    total_stats[mode] = total_stats[mode].set_index('player')

combined_stats = None
elo_combined_stats = None
for mode in total_stats:
    mode_stats = total_stats[mode][['elo']].copy()
    mode_stats[mode] = range(1, len(mode_stats) + 1)
    if combined_stats is None:
        # copy mode stats as-is to combined stats at first
        combined_stats = total_stats[mode][['games', 'wins', 'losses']]
        # copy elo to different variable for special handling
        elo_combined_stats = mode_stats
    else:
        # add values of mode to combined stats
        combined_stats = combined_stats.add(total_stats[mode][['games', 'wins', 'losses']], fill_value=0)
        # add elos
        elo_combined_stats = elo_combined_stats.add(mode_stats, fill_value=0)

# fill missing elo values with 1000
for mode in total_stats:
    elo_combined_stats.loc[elo_combined_stats[mode].isnull(), 'elo'] += 1000
elo_combined_stats = elo_combined_stats['elo']

# calculate elo as the sum of mode elos / 3
combined_stats['elo'] = elo_combined_stats / 3
# sort by elo descendingly first and by player name ascendingly second
combined_stats = combined_stats.sort_values(['elo', 'player'], ascending=[False, True])
# set ranks
combined_stats['rank'] = range(1, len(combined_stats) + 1)
combined_stats = combined_stats.reset_index()
# rename player column to name
combined_stats = combined_stats.rename(columns={'player': 'name'})
# calculate win_rate (= wins / losses)
combined_stats['win_rate'] = (combined_stats['wins']/combined_stats['games']).round(3)
# round integral columns
combined_stats[['elo', 'games', 'losses', 'wins']] = combined_stats[['elo', 'games', 'losses', 'wins']].round(0).astype(int)

# recreate output directory
try:
    shutil.rmtree('output')
except:
    pass
os.mkdir('output')
# save total stats to file
with open(os.path.join('output', 'total.yml'), 'w') as file:
    # begin front matter
    file.write('---\n')
    dump(json.loads(combined_stats.to_json(orient='records')), file, default_flow_style=False)
    # end front matter
    file.write('---')

def as_percentage(series1, series2):
    return (((series1/series2.fillna(1)) * 100).round(0).astype(int).astype(str) + '%').apply(lambda x: x.zfill(3))

def export_stats_yaml(mode_name, mode_stats):
    # reset index to allow player column to be exported
    mode_stats = mode_stats.copy().reset_index()
    mode_stats_copy = mode_stats.copy().reset_index()
    
    # rename player column to name
    mode_stats = mode_stats.rename(columns={'player': 'name'})
    # calculate win_rate (= wins / losses)
    mode_stats['win_rate'] = (mode_stats['wins']/mode_stats['games']).round(3)
    # calculate kpd (= frags / deaths)
    mode_stats['kpd'] = (mode_stats['frags']/mode_stats['deaths']).round(3)
    # account for division by zero
    #mode_stats.loc[mode_stats['damage'] == 0, 'damage'] = 1
    # calculate accuracy (= damage_dealt / damage) * 100 + '%'
    mode_stats['accuracy'] = as_percentage(mode_stats['damage_dealt'], mode_stats['damage'])
    # round integral columns
    mode_stats[['elo', 'games', 'losses', 'wins']] = mode_stats[['elo', 'games', 'losses', 'wins']].round(0).astype(int)
    # select columns to export
    mode_stats = mode_stats[['rank', 'name', 'elo', 'games', 'wins', 'losses', 'win_rate', 'frags', 'deaths', 'kpd', 'suicides', 'accuracy']]
    
    # account for division by zero. since we know damage_dealt is less than or equal to damage for
    # any given weapon, we can simply substitute damage for 1 and the results will be 0/1 = 0
    mode_stats_copy.loc[mode_stats_copy['damage'] == 0, 'damage'] = 1
    mode_stats_copy.loc[mode_stats_copy['damage_0'] == 0, 'damage_0'] = 1
    mode_stats_copy.loc[mode_stats_copy['damage_1'] == 0, 'damage_1'] = 1
    mode_stats_copy.loc[mode_stats_copy['damage_2'] == 0, 'damage_2'] = 1
    mode_stats_copy.loc[mode_stats_copy['damage_4'] == 0, 'damage_4'] = 1
    mode_stats_copy.loc[mode_stats_copy['damage_5'] == 0, 'damage_5'] = 1
    mode_stats_copy.loc[mode_stats_copy['damage_3'] == 0, 'damage_3'] = 1
    mode_stats_copy.loc[mode_stats_copy['damage_6'] == 0, 'damage_6'] = 1
    # calculate accuracy for every weapon (= damage_dealt / damage) * 100 + '%'
    mode_stats['accuracy'] = as_percentage(mode_stats_copy['damage_dealt'], mode_stats_copy['damage'])
    if mode_name != 'insta':
        mode_stats['shotgun'] = as_percentage(mode_stats_copy['damage_dealt_1'], mode_stats_copy['damage_1'])
        mode_stats['chaingun'] = as_percentage(mode_stats_copy['damage_dealt_2'], mode_stats_copy['damage_2'])
        mode_stats['rocket_launcher'] = as_percentage(mode_stats_copy['damage_dealt_3'], mode_stats_copy['damage_3'])
        mode_stats['grenade_launcher'] = as_percentage(mode_stats_copy['damage_dealt_5'], mode_stats_copy['damage_5'])
    if mode_name == 'ffa':
        mode_stats['pistol'] = as_percentage(mode_stats_copy['damage_dealt_6'], mode_stats_copy['damage_6'])
    mode_stats['rifle'] = as_percentage(mode_stats_copy['damage_dealt_4'], mode_stats_copy['damage_4'])
    mode_stats['chainsaw'] = as_percentage(mode_stats_copy['damage_dealt_0'], mode_stats_copy['damage_0'])
    
    # write result to file
    with open(os.path.join('output', '{}.yml'.format(mode_name)), 'w') as file:
        # begin front matter
        file.write('---\n')
        # yaml
        dump(json.loads(mode_stats.to_json(orient='records')), file, default_flow_style=False)
        # end front matter
        file.write('---')

for mode in total_stats:
    mode_name = mode_names[mode]
    # export mode stats
    export_stats_yaml(mode_name, total_stats[mode])

