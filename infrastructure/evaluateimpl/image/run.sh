#!/bin/bash

set -euo pipefail

export PATH=/root/.local/bin:$PATH

obs_path=${OBS_PATH}
evaluate_type=${EVALUATE_TYPE}
obs_ak=${OBS_AK}
obs_sk=${OBS_SK}
obs_endpoint=${OBS_ENDPOINT}
obs_bucket=${OBS_BUCKET}

# workspace
dir=$(pwd)
work_dir=/workspace
test -d $work_dir || mkdir $work_dir
cd $work_dir

# start

if [ $evaluate_type = "custom" ]; then
    obs_util=${OBS_UTIL_PATH}

    $obs_util config -i=$obs_ak -k=$obs_sk -e=$obs_endpoint

    $obs_util cp obs://$obs_bucket/$obs_path $work_dir -f -r > /dev/null 2>&1

    dst=$work_dir/$(basename $obs_path)
    if [ ! -d "$dst" ]; then
        echo "no $dst"
        exit 1
    fi

    aim up --repo $dst --host 0.0.0.0 --port 8080 --workers 2
fi

if [ $evaluate_type = "standard" ]; then
    python3 $dir/aimTrace.py \
	--log_path $obs_path \
	--repo $work_dir \
	--learning_rate_scope "${LEARNING_RATE_SCOPE}" \
	--batch_size_scope "${BATCH_SIZE_SCOPE}" \
	--momentum_scope "${MOMENTUM_SCOPE}" \
	--obs_ak $obs_ak --obs_sk $obs_sk \
	--obs_bucketname $obs_bucket --obs_endpoint $obs_endpoint

    aim up --repo $work_dir --host 0.0.0.0 --port 8080 --workers 2
fi

echo "unknown evaluate type"

exit 1
