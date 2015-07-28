from fabric.api import local, env, sudo

env.hosts = ['nkhumphreys.co.uk']
env.user = 'root'

NAME = "gobananas"


def deploy():
    base_cmd = "scp -r {local_path} root@{host}:{remote_path}"

    remote_path = "/tmp"
    template_path = "/var/www/templates/"
    static_path = "/var/www/static/"

    for h in env.hosts:
        cmd = base_cmd.format(local_path=NAME,
                              host=h,
                              remote_path=remote_path)
        local(cmd)
        cmd = base_cmd.format(local_path="./templates/*",
                              host=h,
                              remote_path=template_path)
        local(cmd)
        cmd = base_cmd.format(local_path="./static/*",
                              host=h,
                              remote_path=static_path)
        local(cmd)

    sudo("mv %s/%s /usr/bin" % (remote_path, NAME))
    sudo("supervisorctl restart %s" % NAME)


def logs():
    cmd = "tail -f /var/log/supervisor/{name}-*.log"
    cmd = cmd.format(name=NAME)
    sudo(cmd)
