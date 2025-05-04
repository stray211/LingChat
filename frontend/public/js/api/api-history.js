import { api } from "./api.js"
import { url } from "./config.js"
export const apiHistory = {
    read: function(data) {return api.get(url.history, data)},
    write: function(data) {return api.post(url.history, data)}
};