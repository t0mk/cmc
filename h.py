#!/usr/bin/env python3

from bs4 import BeautifulSoup

ht = ""
with open("../ih.hmtl") as f:
    ht = f.read()

soup = BeautifulSoup(ht, 'html.parser')

def sf(t):
    #return t.has_attr('class') and 'operation-api-url-path' in t['class']
    return t.name == "operation"

l = soup.find_all(sf)

def surl(t):
    return t.has_attr('class') and 'operation-api-url-path' in t['class']

def spl(t):
    return t.name == 'params-list'

def srl(t):
    return t.name == 'response-list'

def sps(t):
    return t.has_attr('class') and 'param' in t['class']

def spns(t):
    return t.has_attr('class') and 'param-name-wrap' in t['class']

def spts(t):
    return t.has_attr('class') and 'param-type' in t['class']

def spds(t):
    return t.has_attr('class') and 'param-description' in t['class']

def spdefa(t):
    return t.has_attr('class') and 'param-default' in t['class']
# param-default param-enum
d = {}
for i in l:
    ll = i.find_all(surl)
    spla = i.find_all(spl)
    spl0 = spla[0]
    srl0 = i.find_all(srl)
    nam = ll[0].text
    ##print(nam, ":")
    d[nam] = []
            #, 'response': {}}
    for p in spl0.find_all(sps):
        n = p.find_all(spns)[0].get_text(strip=True)
        t = p.find_all(spts)[0].get_text(strip=True)
        desc = p.find_all(spds)[0].get_text(strip=True)
        defa = p.find_all(spdefa)
        pard = dict(name=n, type=t, desc=desc)
        if len(defa) == 1:
            pard['defa'] = defa[0].get_text(strip=True)
        d[nam].append(pard)

import json

out = json.dumps(d, indent="  ")

print(out)
