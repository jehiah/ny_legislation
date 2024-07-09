#!/bin/bash
set -e

mkdir -p build
go build -o build ./sync_ny_legislation

if [ "$(git config --get user.email)" == "" ]; then
    git config --global user.email "automatic-data-update@jehiah.cz"
    git config --global user.name "Data Update"
fi

# git remote rm origin
# git remote add origin https://jehiah:$GH_TOKEN@github.com/jehiah/ny_legislation.git

./build/sync_ny_legislation --target-dir=.


git add bills last_sync.json
git status

FILES_CHANGED=$(git diff --staged --name-only | wc -l)
echo "FILES_CHANGED: ${FILES_CHANGED}"
# if more than one changed commit it (last_sync.json is always updated)
if [[ "${FILES_CHANGED}" -gt 1 ]]; then
    DT=$(TZ=America/New_York date "+%Y-%m-%d %H:%M")
    git commit -a -m "sync: ${DT}"
    git push origin master
fi
