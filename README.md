# CloudStax Amazone EC2 Container Service RPM

This is a fork of [Amazon ECS Init](https://github.com/aws/amazon-ecs-init) to support the
[CloudStax Amazon ECS Container Agent](http://github.com/cloudstax/amazon-ecs-agent), which
adds the volume driver patch.

The original ECS Init caches and loads the ECS Agent tar, which is stored in a S3 bucket. While,this fork directly works with the ECS Agent container image on docker hub.

Other usages are the same with the original Amazon ECS Init. For the details, check [Amazon ECS Init](https://github.com/aws/amazon-ecs-init).

## License

The CloudStax Amazon EC2 Container Service RPM is licensed under the Apache 2.0 License.
