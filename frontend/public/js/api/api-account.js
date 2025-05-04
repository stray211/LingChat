import { api } from "./api.js"
const url = ""
export const apiAccount = {
    login: function(data) {return api.post(url, data)},
    logout: function(data) {return api.post(url, data)},
    register: function(data) {return api.post(url, data)},
    delete: function(data) {return api.post(url, data)},
    passward: function(data) {return api.post(url, data)},
};