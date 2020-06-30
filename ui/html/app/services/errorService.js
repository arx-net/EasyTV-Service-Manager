angular.module("arx-net")
.service("errorService",["$q","modalService","accountService",function($q,modalService,accountService){
    this.handle_error_with_promise = function (error_code){
        no_valid_session(error_code);
        error = {
            title: "error",
            buttons: modalService.buttons["ok"]
        }
        if(error_code != undefined){
            error.message = error_code;
        }
        return $q(function(resolve,reject){
            reject({ data:error });
        });
    }
    this.handle_error = function (error_code){
        no_valid_session(error_code);
        error = {
            title: "error",
            buttons: modalService.buttons["ok"]
        }
        if(error_code != undefined){
            error.message = error_code;
        }
        return error;
    }
    function no_valid_session(error_code){
        if(error_code == -401){
            accountService.logout();
        }
    }
    return this;
}]);