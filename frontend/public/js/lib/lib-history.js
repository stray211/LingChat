import { apiHistory } from "../api/api-history.js"
export const history = {
    read: function() {return apiHistory.read({})},
    export: async function() {
      let data = await apiHistory.read({})
      data = JSON.stringify(data)
      const blob = new Blob([data], { type: "application/json" });
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = "history.txt";
      a.click();
      URL.revokeObjectURL(url);
    },
    import: function(input) {
      const file = input.files[0];
      if (!file) {return};
      const data = new FormData();
      data.append('file',file);
      apiHistory.write(data);
    }
};