import { api } from "./api.js"
import { url } from "./config.js"

export const apiAccount = {
    login: function(data) {return api.post(url.account, data)},
    register: function(data) {return api.post(url.account, data)},
    repassward: function(data) {return api.post(url.account, data)},
    logout: function(data) {return api.post(url.account, data)},
    setting: function(data) {return api.post(url.account, data)},
    delete: function(data) {return api.post(url.account, data)},
    list: function(data) {return api.post(url.account, data)},
    create: function(data) {return api.post(url.account, data)},
};