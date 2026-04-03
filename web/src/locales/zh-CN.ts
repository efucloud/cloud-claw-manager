import menu from "./zh-CN/menu";
import model from "./zh-CN/model";
import model_account from "./zh-CN/model_account";
import pages from "./zh-CN/pages";
export default {
  ...model_account,
  ...pages,
  ...menu,
  ...model,
};
