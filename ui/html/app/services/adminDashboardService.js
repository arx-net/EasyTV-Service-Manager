angular.module("arx-net")
.service("adminDashboardService",["requestService","errorService",function(requestService,errorService){
    this.token = window.sessionStorage.getItem("token");
    this.register_user = function(name,username,email){
        if(name != "" && username != "" && email != ""){
            if(/^\w+([\.-]?\w+)*@\w+([\.-]?\w+)*(\.\w{2,3})+$/.test(email)){
                var data = {
                    name: name,
                    username: username,
                    email: email
                }
                return requestService.post_request("/adm/user/register",JSON.stringify(data),this.token).then(function(response){
                    return response;
                });
            }else{
                return errorService.handle_error_with_promise("client-101");
            }
        }else{
            return errorService.handle_error_with_promise("client-100");
        }
    }
    this.create_service = function(name,description){
        if(name != "" && description != ""){
            var data = {
                name: name,
                description: description
            }
            return requestService.post_request("/adm/service",JSON.stringify(data),this.token).then(function(response){
                return response;
            });
        }else{
            return errorService.handle_error_with_promise("client-100");
        }
    }
    this.get_services = function(){
        return requestService.get_request("/adm/service",this.token).then(function(response){
            return response;
        });
    }
    this.change_password = function(data){
        if(data.old_password != "" && data.new_password != "" && data.new_password_verification != ""){
            if(data.new_password == data.new_password_verification){
                return requestService.post_request("api/user/change_password",JSON.stringify(data),this.token).then(function(response){
                    return response;
                });
            }else{
                return errorService.handle_error_with_promise("client-104");
            }
        }else{
            return errorService.handle_error_with_promise("client-100");
        }
    }
    this.service_availability_change = function(service){
        if(service != ""){
            var data = {
                enable: service.enabled
            }
            return requestService.put_request("/adm/service/"+ service.id,JSON.stringify(data),this.token).then(function(response){
                return response;
            });
        }else{
            return errorService.handle_error_with_promise("client-100");
        }
    }
    return this;
}]);