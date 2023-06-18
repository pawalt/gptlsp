#!/bin/bash

cd $1

while read -r line; do
    if [[ "$line" =~ \[submodule ]]; then
        reading_submodule=1
    elif [[ "$reading_submodule" == 1 ]]; then
        if [[ "$line" =~ path ]]; then
            path=$(echo "$line" | cut -d '=' -f 2 | xargs)
        elif [[ "$line" =~ url ]]; then
            url=$(echo "$line" | cut -d '=' -f 2 | xargs)
        fi

        if [[ -n "$path" && -n "$url" ]]; then
            git clone "$url" "$path"
            unset path
            unset url
            reading_submodule=0
        fi
    fi
done < .gitmodules