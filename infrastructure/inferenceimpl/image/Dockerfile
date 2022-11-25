FROM alpine:3.14 as BUILDER
RUN apk update && apk add --no-cache \
        git \
        bash \
        libc6-compat
WORKDIR /opt/build
COPY ./obsutil.tar.gz ./obsutil.tar.gz
RUN tar -xf ./obsutil.tar.gz

# real service
FROM python:3.9.13

USER root
RUN groupadd --gid 5000 mindspore \
  && useradd --home-dir /home/mindspore --create-home --uid 5000 --gid 5000 --shell /bin/sh --skel /dev/null mindspore

RUN apt-get update
RUN apt-get install ffmpeg libsm6 libxext6 -y
RUN apt install libgl1-mesa-glx

USER mindspore

WORKDIR /usr/src/app

COPY --chown=mindspore:mindspore ./run.sh ./run.sh
COPY --chown=mindspore:mindspore ./requirements.txt ./requirements.txt
COPY --chown=mindspore:mindspore ./pretrain.py ./pretrain.py
COPY --chown=mindspore:mindspore --from=BUILDER /opt/build/obsutil /usr/src/app/obsutil

RUN pip install --upgrade -i https://pypi.mirrors.ustc.edu.cn/simple pip
RUN pip install --default-timeout=1000 --user -i https://pypi.mirrors.ustc.edu.cn/simple -r requirements.txt

# TODO: what's the purpose
EXPOSE 7860

ENTRYPOINT ["/usr/src/app/run.sh"]
