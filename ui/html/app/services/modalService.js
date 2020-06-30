app
.service("modalService",["ngDialog",function(ngDialog){
    this.modals = {
        dialog: "app/components/modal/dialog.html",
        popup: "app/components/modal/popup.html",
        task_list: "app/components/modal/task_list.html",
        output: "app/components/modal/output.html"
    }
    this.buttons = {
        "ok/cancel" : [
            {
                name: "ok",
                classes: "btn sm-button mx-1",
                return: 1
            },
            {
                name:"cancel",
                classes: "btn btn-danger mx-1",
                return: 0
            }
        ],
        "yes/no" : [
            {
                name: "yes",
                classes: "btn sm-button mx-1",
                return: 1
            },
            {
                name:"no",
                classes: "btn btn-danger mx-1",
                return: 0
            }
        ],
        "ok": [
            {
                name: "ok",
                classes: "btn sm-button",
                return: 1
            }
        ],
        "cancel": [
            {
                name:"cancel",
                classes: "btn btn-danger",
                return: 0
            }
        ]
    }
    this.open_dialog = function(template,data,func,$scope,controller){
        var dialog = ngDialog.open({ 
            template: template, 
            className: 'ngdialog-theme-default',
            controller: controller,
            closeByEscape: true,
            scope:$scope,
            data: data,
            preCloseCallback: function(ok){
                if(func != undefined){
                    func(ok);
                } 
            } 
        });
    }
    this.open_popup = function(template, data, width){
        ngDialog.open({ 
            template: template, 
            className: 'ngdialog-theme-default',
            closeByEscape: true,
            data: data,
            width: width
        });
    }
    return this;
}]);