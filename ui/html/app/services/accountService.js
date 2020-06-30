angular.module("arx-net")
.service("accountService",["requestService",function(requestService){
    this.logout = function(language){
        var token = window.sessionStorage.getItem("token");

        requestService.delete_request("/api/user/logout", token)
        .then(function(response){
            console.dir(response);
            window.sessionStorage.removeItem("token");
            window.sessionStorage.removeItem("is_admin");
            window.location.hash = "";
            window.location.pathname = "/index.html";
        });
    }
    return this;
}]);