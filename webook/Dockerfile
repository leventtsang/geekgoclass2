# 基础镜像
FROM ubuntu:22.04
# Set Aliyun sources for apt
RUN echo "deb http://mirrors.aliyun.com/ubuntu/ focal main restricted universe multiverse" > /etc/apt/sources.list && \
    echo "deb http://mirrors.aliyun.com/ubuntu/ focal-security main restricted universe multiverse" >> /etc/apt/sources.list && \
    echo "deb http://mirrors.aliyun.com/ubuntu/ focal-updates main restricted universe multiverse" >> /etc/apt/sources.list && \
    echo "deb http://mirrors.aliyun.com/ubuntu/ focal-proposed main restricted universe multiverse" >> /etc/apt/sources.list && \
    echo "deb http://mirrors.aliyun.com/ubuntu/ focal-backports main restricted universe multiverse" >> /etc/apt/sources.list

# Install prerequisites including gcc and other build essentials
#RUN apt-get update && apt-get install -y build-essential

# 把编译后的打包进来这个镜像，放到工作目录 /app
COPY webook /app/webook
WORKDIR /app
ENTRYPOINT ["/app/webook"]
