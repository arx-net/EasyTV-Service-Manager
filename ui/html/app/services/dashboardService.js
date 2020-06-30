angular.module("arx-net")
.service("dashboardService",["requestService","errorService","$http",function(requestService,errorService,$http){
    this.token = window.sessionStorage.getItem("token");
    this.get_services = function(){
        return requestService.get_request("/api/service",this.token).then(function(response){
            return response;
            console.dir(response);
        });
    }
    this.get_service = function(id){
        return requestService.get_request("/api/service/"+id,this.token).then(function(response){
            console.dir(response);
            return response;
        });
    }
    this.get_jobs = function(){
        return  requestService.get_request("/api/job",this.token).then(function(response){
            console.dir(response);
            return response
        });
    }
    this.get_job = function(){
        return  requestService.get_request("/api/job",this.token).then(function(response){
            return response
        });
    }
    this.create_job = function(public_date,exp_date,tasks){
        if(public_date != NaN && exp_date != NaN && tasks.length > 0){
            var error = check_tasks_inputs(tasks);
            if(!error){
                var job = {
                    publication_date : public_date,
                    expiration_date: exp_date,
                    tasks: tasks
                }
                return requestService.post_request("/api/job",job,this.token).then(function(response){
                    console.dir(response);
                    return response;
                });
            }else{
                return errorService.handle_error_with_promise("client-100");
            }
        }else{
            return errorService.handle_error_with_promise("client-100");
        }
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
    this.cancel_job = function(id){
        return requestService.delete_request("api/job/"+id,this.token).then(function(response){
            console.dir(response);
            return response
        })
    }
    function check_tasks_inputs(tasks){
        for(var i in tasks){
            if(tasks[i].input != undefined){
                for(var key in tasks[i].input){
                    if(tasks[i].input[key] == undefined){
                        return true;
                    }
                }
            }else{
                return true;
            }
        }
        return false;
    }
    window.enable_tasks = function(){
        $http({
            method: "PUT",
            url: "/internal/task/1",
            headers:{
                "X-EasyTV-Session" : this.token,
                "X-EasyTV-Key": "f266b6c1d209c2d68fb0c0b523773d0aa1dfc4a527207a5b199f11b42936396e"
            },
            data:{
                "disabled": false
            }
        }).then(function successCallback(response) {
            // this callback will be called asynchronously
            // when the response is available
            return response;
        }, function errorCallback(response) {
            // called asynchronously if an error occurs
            // or server returns response with an error status.
            return response;
        }).then(function(response){
            console.dir(response);
        })
    }
    window.create_tasks = function(){
        var task = {
            "name": "paragogi ipotitlwn batman",
            "description": "paragogi ipotitlwn",
            "start_url": "https://service/start_job",
            "cancel_url": "https://service/start_job",
            "input": {
                "language_source": "string",
                "language_export": "string"
            },
            "output": {
                "language": "string"
            }
        }
        $http({
            method: "POST",
            url: "/internal/task",
            headers:{
                "X-EasyTV-Session" : this.token,
                "X-EasyTV-Key": "f266b6c1d209c2d68fb0c0b523773d0aa1dfc4a527207a5b199f11b42936396e"
            },
            data:task
        }).then(function successCallback(response) {
            // this callback will be called asynchronously
            // when the response is available
            return response;
        }, function errorCallback(response) {
            // called asynchronously if an error occurs
            // or server returns response with an error status.
            return response;
        }).then(function(response){
            console.dir(response);
        })
    }
    return this;
}]);