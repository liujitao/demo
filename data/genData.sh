#!/bin/bash

for i in {10..19}
do
    id=$(uuid)
    name=$(echo user$i)
    mobile=$(echo 139000000$i)
    email=$(echo user$i@abc.com)
    let j=i-10
    create_time=$(echo 2021-12-30T10:00:0$j.000Z)

    echo {\"uuid\": \"$id\", \"user_name\": \"$name\", \"real_name\": \"$name\", \"mobile\": \"$mobile\", \"email\": \"$email\", \"password\": \"password\", \"create_at\": \"$create_time\", \"update_at\": \"$create_time\"}, >> userData.json
done

for i in {20..69}
do
    id=$(uuid)
    name=$(echo user$i)
    mobile=$(echo 139000000$i)
    email=$(echo user$i@abc.com)
    let j=i-10
    create_time=$(echo 2021-12-30T11:00:$j.000Z)

    echo {\"uuid\": \"$id\", \"user_name\": \"$name\", \"real_name\": \"$name\", \"mobile\": \"$mobile\", \"email\": \"$email\", \"password\": \"password\", \"create_at\": \"$create_time\", \"update_at\": \"$create_time\"}, >> userData.json
done

for i in {70..99}
do
    id=$(uuid)
    name=$(echo user$i)
    mobile=$(echo 139000000$i)
    email=$(echo user$i@abc.com)
    let j=i-40
    create_time=$(echo 2021-12-30T12:00:$j.000Z)

    echo {\"uuid\": \"$id\", \"user_name\": \"$name\", \"real_name\": \"$name\", \"mobile\": \"$mobile\", \"email\": \"$email\", \"password\": \"password\", \"create_at\": \"$create_time\", \"update_at\": \"$create_time\"}, >> userData.json
done

exit 0