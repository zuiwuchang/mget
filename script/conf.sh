Target="mget"
Docker="github.com/zuiwuchang/mget"
Dir=$(cd "$(dirname $BASH_SOURCE)/.." && pwd)
Version="v0.0.1"
Platforms=(
    darwin/amd64
    windows/amd64
    linux/arm
    linux/amd64
)
