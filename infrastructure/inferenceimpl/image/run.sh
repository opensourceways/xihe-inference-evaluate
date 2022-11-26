#!/bin/bash

set -euo pipefail

pretrain_file=$(pwd)/pretrain.py

gitlab_endpoint=${GITLAB_ENDPOINT}
xihe_user=${XIHE_USER}
xihe_user_token=${XIHE_USER_TOKEN}
project_name=${PROJECT_NAME}
last_commit=${LAST_COMMIT}

repo_url=http://${xihe_user}:${xihe_user_token}@${gitlab_endpoint#"http://"}

obs_ak=${OBS_AK}
obs_sk=${OBS_SK}
obs_endpoint=${OBS_ENDPOINT}
obs_util=${OBS_UTIL_PATH}
obs_lfs_path=${OBS_BUCKET}/${OBS_LFS_PATH} #$OBS_LFS_PATH has no suffix of /.

$obs_util config -i=$obs_ak -k=$obs_sk -e=$obs_endpoint

app=app.py
work_dir=/workspace
inference_dir=$work_dir/$project_name/inference

# helper
download_model() {
    local owner=$1
    local repo=$2
    local file=$3

    cd $work_dir

    test -d $owner || mkdir $owner

    cd $owner

    if [ ! -d $repo ]; then
        git clone $repo_url/${owner}/${repo}
    fi

    if [ ! -e $file ]; then
        echo "no model file: $file"
        exit 1
    fi

    # check if lfs
    sha=$(sed -n '/^oid sha256:.\{64\}$/p' "$file")
    if [ -n "$sha" ]; then
        sha=${sha#"oid sha256:"}
        dst=$inference_dir/$(basename $file)

        $obs_util cp obs://$obs_lfs_path/${sha:0:2}/${sha:2:2}/${sha:4} $dst > /dev/null 2>&1

        if [ ! -e "$dst" ]; then
            echo "no $dst"
            exit 1
        fi
    else
        # small file
        mv $file $inference_dir
    fi
}

# workspace
test -d $work_dir || mkdir $work_dir
cd $work_dir

# download project
git clone $repo_url/$xihe_user/$project_name

if [ ! -e "$inference_dir/$app" ]; then
    echo "no $app"
    exit 1
fi

# download dependents
f="$inference_dir/config.json"

if [ -e "$f" -a -s "$f" ]; then
    mf=$work_dir/dependent_models

    python3 $pretrain_file $f > $mf 2>&1
    if [ $? -ne 0 ]; then
        echo $(cat $mf)
        exit 1
    fi

    while read line;
    do
        download_model $line
    done < $mf
fi

# run
cd $inference_dir

f=requirements.txt
if [ -e "$f" -a -s "$f" ]; then
    pip install --upgrade -i https://pypi.tuna.tsinghua.edu.cn/simple pip
    pip install -r ./$f -i https://pypi.tuna.tsinghua.edu.cn/simple
fi

python3 ./$app
