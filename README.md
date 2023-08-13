
#1、login接口采用jwt-token，暂时还是只收集邮箱密码信息。
![Alt text](image.png)

#2、完善了一下/users/profile接口，能返回相关个人信息，并排除了password字段。
![Alt text](image-2.png)

#3、/users/edit接口目前能修改昵称、生日、简介，通过/users/profile接口能重新查询。
![Alt text](image-3.png)
![Alt text](image-4.png)
![Alt text](image-5.png)