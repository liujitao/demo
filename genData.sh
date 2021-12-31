#!/bin/bash

for i in {0..9}
do
    id=$(echo 61cd47b338e4acbd43065c0$i)
    name=$(echo user$i)
    mobile=$(echo 1390000000$i)
    email=$(echo user$i@abc.com)
    create_time=$(echo 2021-12-30T09:00:0$i.000Z)

    echo {\"_id\": \"$id\", \"user_name\": \"$name\", \"real_name\": \"$name\", \"mobile\": \"$mobile\", \"email\": \"$email\", \"password\": \"password\", \"create_at\": \"$create_time\", \"update_at\": \"$create_time\"}, >> userData.json
done

for i in {10..19}
do
    id=$(echo 61cd47b338e4acbd43065c$i)
    name=$(echo user$i)
    mobile=$(echo 139000000$i)
    email=$(echo user$i@abc.com)
    let j=i-10
    create_time=$(echo 2021-12-30T10:00:0$j.000Z)

    echo {\"_id\": \"$id\", \"user_name\": \"$name\", \"real_name\": \"$name\", \"mobile\": \"$mobile\", \"email\": \"$email\", \"password\": \"password\", \"create_at\": \"$create_time\", \"update_at\": \"$create_time\"}, >> userData.json
done

for i in {20..69}
do
    id=$(echo 61cd47b338e4acbd43065c$i)
    name=$(echo user$i)
    mobile=$(echo 139000000$i)
    email=$(echo user$i@abc.com)
    let j=i-10
    create_time=$(echo 2021-12-30T11:00:$j.000Z)

    echo {\"_id\": \"$id\", \"user_name\": \"$name\", \"real_name\": \"$name\", \"mobile\": \"$mobile\", \"email\": \"$email\", \"password\": \"password\", \"create_at\": \"$create_time\", \"update_at\": \"$create_time\"}, >> userData.json
done

for i in {70..99}
do
    id=$(echo 61cd47b338e4acbd43065c$i)
    name=$(echo user$i)
    mobile=$(echo 139000000$i)
    email=$(echo user$i@abc.com)
    let j=i-40
    create_time=$(echo 2021-12-30T12:00:$j.000Z)

    echo {\"_id\": \"$id\", \"user_name\": \"$name\", \"real_name\": \"$name\", \"mobile\": \"$mobile\", \"email\": \"$email\", \"password\": \"password\", \"create_at\": \"$create_time\", \"update_at\": \"$create_time\"}, >> userData.json
done

exit 0