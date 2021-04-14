drop database if exists sales_storage;
create database sales_storage;
use sales_storage;

-- 商品仓库或者门店表
drop table if exists goods_location;
create table goods_location(
     id int auto_increment primary key,
     name varchar(128) not null,
     address varchar(512) comment '门店或者仓库地址',
     update_time datetime default now()
);


-- 商品信息表
drop table if exists goods_info;
create table goods_info(
    id int auto_increment primary key,
    goods_name varchar(128) not null,
    bar_code varchar(128) not null,
    sales_price decimal(30,6) not null comment '规定销售价格',
    update_time datetime default now()
);
-- 入库表
drop table if exists goods_input;
create table goods_input(
    id int auto_increment primary key,
    batch_no varchar(32) not null comment '批次号',
    storage_id int not null comment '对应goods_storage.id',
    goods_id int not null comment 'reference id to id in table goods_info',
    expire_date date not null,
    quantity decimal(30,2) not null,
    cost decimal(30,6) not null comment '进货价',
    update_time datetime default now()
);

-- 商品库存表
drop table if exists goods_storage;
create table goods_storage(
    id int auto_increment primary key,
    goods_id int not null comment 'reference id to id in table goods_info',
    quantity decimal(30,2) not null,
    price decimal(30,6) not null comment 'average price',
    update_time datetime default now()
);

-- 商品location修改日志表
drop table if exists goods_location_change_log;
create table goods_location_change_log(
    id int auto_increment primary key,
    goods_id int not null,
    from_location int  comment '商品所在仓库或者门店',
    to_location int  comment '商品所在仓库或者门店',
    update_time datetime default now()
);


-- 商品销售日志表
drop table if exists goods_sales_log;
create table goods_sales_log(
    id int auto_increment primary key,
    goods_id int not null,
    sales_location int  comment '商品所在仓库或者门店',
    sales_price decimal(30,6) comment '实际销售价格',
    sales_quantity decimal(30,6),
    storage_id int ,
    update_time datetime default now()
);