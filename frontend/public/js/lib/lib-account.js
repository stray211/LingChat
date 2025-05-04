import { apiAccount } from "../api/api-account.js"

function identity(data) {
  data.accountId = "id"
  return data
}

function parse(data, form) {
  const formData = new FormData(form);
  for (const [identity ,value] of formData.entries()) {
    data[identity].push(value)
  };
  return data
}

export let account = {
  login: function(form) {
    let data = {};
    data = identity(data);
    data = parse(data, form);
    return apiAccount.login(data);
  },
  register: function(form) {
    let data = {};
    data = identity(data);
    data = parse(data, form);
    return apiAccount.register(data);
  },
  repassword: function(form) {
    let data = {};
    data = identity(data);
    data = parse(data, form);
    return apiAccount.repassword(data);
  },
  setting: function(form) {
    let data = {};
    data = identity(data);
    data = parse(data, form);
    return apiAccount.setting(data);
  },
  logout: function() {
    let data = {};
    data = identity(data);
    return apiAccount.logout(data);
  },
  list: function() {
    let data = {};
    data = identity(data);
    return apiAccount.list(data);
  },
  create: function() {
    let data = {};
    data = identity(data);
    return apiAccount.create(data);
  },
  delete: function() {
    let data = {};
    data = identity(data);
    return apiAccount.delete(data);
  },
};