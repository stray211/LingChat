import { apiAccount } from "../api/api-account.js"

function parse(form) {
  const data = {}
  const formData = new FormData(form);
  for (const [key, value] of formData.entries()) {
    data[key].push(value)
  };
  return data
}

export let account = {
  login: function(form) {return apiAccount.login(parse(form))},
  register: function(form) {return apiAccount.register(parse(form))},
  repassword: function(form) {return apiAccount.repassword(parse(form))},
  setting: function(form) {return apiAccount.setting(parse(form));},
  logout: function() {return apiAccount.logout({})},
  list: function() {return apiAccount.list({})},
  check: function(id) {return apiAccount.check({id: id})},
  delete: function(id) {return apiAccount.delete({id: id})},
};