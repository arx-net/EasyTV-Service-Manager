angular.module("arx-net")
.service("requestService",["$http",function($http){
    this.post_request = function (url,json,token){
        return $http({
            method: "POST",
            url: url,
            headers:{
                "X-EasyTV-Session" : token
            },
            data:json
        }).then(function successCallback(response) {
            // this callback will be called asynchronously
            // when the response is available
            return response;
        }, function errorCallback(response) {
            // called asynchronously if an error occurs
            // or server returns response with an error status.
            return response;
        });
    }
    this.get_request = function (url,token){
        return $http({
            method: "GET",
            url: url,
            headers:{
                "X-EasyTV-Session" : token
            }
        }).then(function successCallback(response) {
            // this callback will be called asynchronously
            // when the response is available
            return response;
          }, function errorCallback(response) {
            // called asynchronously if an error occurs
            // or server returns response with an error status.
            return response;
          });
    }
    this.delete_request = function(url,token){
        return $http({
            method:"DELETE",
            url: url,
            headers:{
                "X-EasyTV-Session" : token
            }
        }).then(function successCallback(response) {
            // this callback will be called asynchronously
            // when the response is available
            return response;
          }, function errorCallback(response) {
            // called asynchronously if an error occurs
            // or server returns response with an error status.
            return response;
          });
    }
    this.put_request = function(url,json,token){
        return $http({
            method:"PUT",
            url: url,
            headers:{
                "X-EasyTV-Session" : token
            },
            data:json
        }).then(function successCallback(response) {
            // this callback will be called asynchronously
            // when the response is available
            return response;
        }, function errorCallback(response) {
            // called asynchronously if an error occurs
            // or server returns response with an error status.
            return response;
        });
    }
    return this;
}]);