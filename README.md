# CloudStax Amazone EC2 Container Service RPM

This is a fork of [Amazon ECS Init](https://github.com/aws/amazon-ecs-init) to support the
[CloudStax Amazon ECS Container Agent](http://github.com/cloudstax/amazon-ecs-agent), which
adds the volume driver patch.

The original ECS Init caches and loads the ECS Agent tar, which is stored in a S3 bucket. While,this fork directly works with the ECS Agent container image on docker hub.

Other usages are the same with the original Amazon ECS Init. For the details, check [Amazon ECS Init](https://github.com/aws/amazon-ecs-init).

Simply make to build the cloudstax-ecs-init rpm. Note: the RPM MUST be built on the Amazon Linux instance. The RPM built on Ubuntu could not run on Amazon Linux instance.

Steps:
- install Amazon Linux AMI
- sudo yum install docker
- sudo yum install golang
- sudo yum install rpm-build
- go get github.com/cloudstax/amazon-ecs-init
- make rpm

## License

The CloudStax Amazon EC2 Container Service RPM is licensed under the Apache 2.0 License.
