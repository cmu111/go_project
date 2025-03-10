package interceptor

import (
	"context"
	"encoding/json"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"test.com/project-common/encrypts"
	"test.com/project-grpc/user/login"
	"test.com/project-user/internal/dao"
	"test.com/project-user/internal/repo"
)

// 创建结构体，含有map成员变量，存路由到返回结构体对象的映射
// 成员函数返回grpc.ServerOption，使grpc.server可以调用该接口的apply方法来注册拦截器
// 成员函数中调用grpc.UnaryInterceptor(),返回一个实现了grpc.ServerOption接口的对象，该对象的成员变量就是此处写的函数。该对象的apply就是使用其成员变量（该处的函数）将拦截器注册到grpc服务中
type CacheInterceptor struct {
	cacheMap map[string]any
	cache    repo.Cache
}

func NewCacheInterceptor() *CacheInterceptor {
	cacheMap := make(map[string]any)
	cacheMap["/login.service.v1.LoginService/MyOrgList"] = &login.OrgListResponse{}
	cacheMap["/login.service.v1.LoginService/FindMemInfoById"] = &login.MemberMessage{}
	return &CacheInterceptor{
		cacheMap: cacheMap,
		cache:    dao.Rc,
	}
}

// 调用cache函数返回一serverOption对象，封装好了拦截器，newServer时传入该对象来注册拦截器
func (c *CacheInterceptor) Cache() grpc.ServerOption {
	return grpc.UnaryInterceptor(func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		reqType := c.cacheMap[info.FullMethod]
		if reqType == nil {
			return handler(ctx, req)
		}
		//如果有数据，直接获取返回
		con, cancel := context.WithTimeout(ctx, time.Second*5)
		defer cancel()
		//key使用请求的md5值，保证请求的实例相同时才命中缓存
		marshal, _ := json.Marshal(req)
		cacheKey := encrypts.Md5(string(marshal))
		value, err := c.cache.Get(con, info.FullMethod+"::"+cacheKey)
		if value != "" {
			json.Unmarshal([]byte(value), &reqType)
			zap.L().Info(info.FullMethod + " 走了缓存")
			return reqType, nil
		}
		//如果没有数据，则调用handler处理请求，并将返回值存入缓存
		resp, err = handler(ctx, req)
		if err != nil {
			return nil, err
		}
		marshal, _ = json.Marshal(resp)
		c.cache.Put(con, info.FullMethod+"::"+cacheKey, string(marshal), 5*time.Minute)
		return resp, nil
	})
}
