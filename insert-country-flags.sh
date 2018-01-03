#!/bin/env bash

sed -i -E \
    -e "s/(Redon)/<img class=\"flag\" src=\"http:\/\/sauerduels.me\/images\/nl.png\" \/> \1/g" \
    -e "s/(Honzik1)/<img class=\"flag\" src=\"http:\/\/sauerduels.me\/images\/cz.png\" \/> \1/g" \
    -e "s/(Lokio)/<img class=\"flag\" src=\"http:\/\/sauerduels.me\/images\/us.png\" \/> \1/g" \
    -e "s/(Frosty)/<img class=\"flag\" src=\"http:\/\/sauerduels.me\/images\/ca.png\" \/> \1/g" \
    -e "s/(DarkFire)/<img class=\"flag\" src=\"http:\/\/sauerduels.me\/images\/us.png\" \/> \1/g" \
    -e "s/(Alluro)/<img class=\"flag\" src=\"http:\/\/sauerduels.me\/images\/de.png\" \/> \1/g" \
    -e "s/(Partizan)/<img class=\"flag\" src=\"http:\/\/sauerduels.me\/images\/rs.png\" \/> \1/g" \
    -e "s/(BudSpencer)/<img class=\"flag\" src=\"http:\/\/sauerduels.me\/images\/de.png\" \/> \1/g" \
    -e "s/(starch)/<img class=\"flag\" src=\"http:\/\/sauerduels.me\/images\/us.png\" \/> \1/g" \
    -e "s/(hades)/<img class=\"flag\" src=\"http:\/\/sauerduels.me\/images\/us.png\" \/> \1/g" \
    -e "s/(Rexus)/<img class=\"flag\" src=\"http:\/\/sauerduels.me\/images\/gb.png\" \/> \1/g" \
    -e "s/(lagout)/<img class=\"flag\" src=\"http:\/\/sauerduels.me\/images\/fr.png\" \/> \1/g" \
    -e "s/(Gangler)/<img class=\"flag\" src=\"http:\/\/sauerduels.me\/images\/de.png\" \/> \1/g" \
    -e "s/(bug)/<img class=\"flag\" src=\"http:\/\/sauerduels.me\/images\/il.png\" \/> \1/g" \
    -e "s/(Fear)/<img class=\"flag\" src=\"http:\/\/sauerduels.me\/images\/de.png\" \/> \1/g" \
    -e "s/(Aiurz)/<img class=\"flag\" src=\"http:\/\/sauerduels.me\/images\/ch.png\" \/> \1/g" \
    -e "s/(BenKei)/<img class=\"flag\" src=\"http:\/\/sauerduels.me\/images\/nl.png\" \/> \1/g" \
    -e "s/(Bertolt_Brecht)/<img class=\"flag\" src=\"http:\/\/sauerduels.me\/images\/gb.png\" \/> \1/g" \
    -e "s/(Rudi)/<img class=\"flag\" src=\"http:\/\/sauerduels.me\/images\/za.png\" \/> \1/g" \
    -e "s/(h8)/<img class=\"flag\" src=\"http:\/\/sauerduels.me\/images\/in.png\" \/> \1/g" \
    -e "s/(Khornettoh)/<img class=\"flag\" src=\"http:\/\/sauerduels.me\/images\/fr.png\" \/> \1/g" \
    -e "s/(Tamin0)/<img class=\"flag\" src=\"http:\/\/sauerduels.me\/images\/de.png\" \/> \1/g" \
    -e "s/(ZCrone)/<img class=\"flag\" src=\"http:\/\/sauerduels.me\/images\/fr.png\" \/> \1/g" \
    -e "s/(Josh22)/<img class=\"flag\" src=\"http:\/\/sauerduels.me\/images\/gb.png\" \/> \1/g" \
    -e "s/(greenadiss)/<img class=\"flag\" src=\"http:\/\/sauerduels.me\/images\/ua.png\" \/> \1/g" \
    -e "s/(fatmonkeygenius)/<img class=\"flag\" src=\"http:\/\/sauerduels.me\/images\/us.png\" \/> \1/g" \
    -e "s/(Headway)/<img class=\"flag\" src=\"http:\/\/sauerduels.me\/images\/ru.png\" \/> \1/g" \
    -e "s/(Harsh)/<img class=\"flag\" src=\"http:\/\/sauerduels.me\/images\/us.png\" \/> \1/g" \
    -e "s/(yggdrasil)/<img class=\"flag\" src=\"http:\/\/sauerduels.me\/images\/at.png\" \/> \1/g" \
    -e "s/(Phönix)/<img class=\"flag\" src=\"http:\/\/sauerduels.me\/images\/de.png\" \/> \1/g" \
    -e "s/(NoobGuy)/<img class=\"flag\" src=\"http:\/\/sauerduels.me\/images\/us.png\" \/> \1/g" \
    $1
