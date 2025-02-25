#!/bin/bash

find_default_branch () {
    branch_name=$(curl -sL \
    -H "Accept: application/vnd.github+json" \
    -H "Authorization: Bearer $GH_TOKEN" \
    -H "X-GitHub-Api-Version: 2022-11-28" \
    https://api.github.ibm.com/repos/turbonomic/$1 | jq -r '.default_branch')
    echo $branch_name
}

# find direct internal dependencies from go.mod file
current_repo=prometurbo
turbo_repos=$(cat go.mod | grep github.ibm.com | grep -v indirect | grep -v $current_repo | awk '{ print $1 }')

go env -w GOPRIVATE=github.ibm.com
packages=()
while IFS= read -r repo
do
    package=$(echo "$repo" | awk -F  '/' '{print $NF}')
    default_branch=$(find_default_branch $package)
    echo "Retriving latest package for $package on branch $default_branch"
    go get github.ibm.com/turbonomic/$package@$default_branch
    packages+=("$package")
   
done <<< "$turbo_repos"

# sync vendor directory
go mod vendor

diff_out=$(git diff go.mod)
echo $diff_out
if [ ! -z "$diff_out" ]; then
    err_msg=""
    for p in "${packages[@]}"
    do
        err_msg+=$(echo "$diff_out" | grep ^- | grep $p)
        err_msg+=$(echo "$diff_out" | grep ^+ | grep $p)
    done <<< "$turbo_repos"

    if [ ! -z "$err_msg" ]; then
        echo "New dependencies detected!"
        echo "$err_msg"
        echo "Commit the changes introduced by this script before running the build again!"
        exit 1
    fi
fi

echo "All turbo dependencies are up to date!"