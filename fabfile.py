from fabric.api import local, env, sudo

env.hosts = ['nkhumphreys.co.uk']

NAME = "golang_blog"


def deploy():
    cmd = "scp {local_path} root@{host}:{remote_path}"

    remote_path = "/tmp"

    for h in env.hosts:
        cmd = cmd.format(local_path=NAME,
                         host=h,
                         remote_path=remote_path)
        local(cmd)

    sudo("mv %s/%s /usr/bin" % (remote_path, NAME))
    sudo("supervisorctl restart %s" % NAME)


def logs():
    cmd = "tail -f /var/log/supervisor/{name}-*.log"
    cmd = cmd.format(name=NAME)
    sudo(cmd)
