document.addEventListener("DOMContentLoaded", () => {
  // 为登录页面设置随机背景图片
  setRandomBackgroundImage();

  // DOM Elements
  const loginForm = document.getElementById("login-form");
  const registerForm = document.getElementById("register-form");
  const resetPwdForm = document.getElementById("reset-pwd-form");
  const loginTab = document.getElementById("login-tab");
  const registerTab = document.getElementById("register-tab");
  const loginFormElement = document.getElementById("login-form-element");
  const registerFormElement = document.getElementById("register-form-element");
  const resetPwdFormElement = document.getElementById("reset-pwd-form-element");
  const loginMessage = document.getElementById("login-message");
  const registerMessage = document.getElementById("register-message");
  const resetPwdMessage = document.getElementById("reset-pwd-message");
  const rememberMeCheckbox = document.getElementById("remember-me");
  const showResetPasswordLink = document.getElementById("show-reset-password");
  const cancelResetPasswordBtn = document.getElementById(
    "cancel-reset-password"
  );

  // 确保表单和按钮正常工作
  console.log("登录表单元素:", loginFormElement);
  console.log("注册表单元素:", registerFormElement);
  console.log("登录按钮:", document.getElementById("btn-login"));
  console.log("注册按钮:", document.getElementById("btn-register"));

  // Check if user is already logged in
  checkLoginStatus();

  // Auto-fill login form if remember-me was checked previously
  autoFillLoginForm();

  // Tab switching
  loginTab.addEventListener("click", () => {
    loginTab.classList.add("active");
    registerTab.classList.remove("active");
    loginForm.style.display = "block";
    registerForm.style.display = "none";
    resetPwdForm.style.display = "none";
    loginMessage.textContent = "";
  });

  registerTab.addEventListener("click", () => {
    loginTab.classList.remove("active");
    registerTab.classList.add("active");
    loginForm.style.display = "none";
    registerForm.style.display = "block";
    resetPwdForm.style.display = "none";
    registerMessage.textContent = "";
  });

  // 显示修改密码表单
  showResetPasswordLink?.addEventListener("click", (e) => {
    e.preventDefault();
    loginForm.style.display = "none";
    registerForm.style.display = "none";
    resetPwdForm.style.display = "block";
    resetPwdMessage.textContent = "";

    // 如果已记住用户名，自动填充
    const rememberedUsername = localStorage.getItem("remembered_username");
    if (rememberedUsername) {
      document.getElementById("reset-email").value = rememberedUsername;
    }
  });

  // 取消修改密码
  cancelResetPasswordBtn?.addEventListener("click", () => {
    resetPwdForm.style.display = "none";
    loginForm.style.display = "block";
    loginTab.classList.add("active");
    registerTab.classList.remove("active");
  });

  // Handle login form submission
  loginFormElement.addEventListener("submit", async (e) => {
    e.preventDefault();
    loginMessage.textContent = "";
    loginMessage.className = "form-message";

    const username = document.getElementById("email").value;
    const password = document.getElementById("password").value;
    const rememberMe = rememberMeCheckbox ? rememberMeCheckbox.checked : false;

    // Validate form
    if (!username || !password) {
      loginMessage.textContent = "请填写所有必填字段";
      return;
    }

    // Disable the form during submission
    setFormLoading(loginFormElement, true);

          try {
        console.log("发送登录请求:", { username, rememberMe });

        const response = await fetch("/api/v1/user/login", {
          method: "POST",
          headers: {
            "Content-Type": "application/json",
          },
          body: JSON.stringify({ username, password }),
        });

      const data = await response.json();
      console.log("登录响应:", {
        code: data.code,
        msg: data.msg,
      });

      if (data.code === 200) {
        loginMessage.textContent = "登录成功，正在跳转...";
        loginMessage.className = "form-message success";

        // Remember login info if checkbox is checked
        if (rememberMe) {
          localStorage.setItem("remembered_username", username);
          localStorage.setItem("remembered_password", btoa(password)); // 简单加密，非安全存储
          localStorage.setItem("remember_me", "true");
        } else {
          localStorage.removeItem("remembered_username");
          localStorage.removeItem("remembered_password");
          localStorage.removeItem("remember_me");
        }

        // Save JWT token instead of user object
        JWTUtils.saveToken(data.data.token);

        // Redirect to chat page after a short delay
        setTimeout(() => {
          window.location.href = "/";
        }, 1000);
      } else {
        loginMessage.textContent = data.msg || "登录失败，请检查邮箱和密码";
      }
    } catch (error) {
      console.error("登录错误:", error);
      loginMessage.textContent = "登录时发生错误，请稍后再试";
    } finally {
      setFormLoading(loginFormElement, false);
    }
  });

  // 直接给注册按钮添加点击事件
  const registerButton = document.getElementById("btn-register");
  if (registerButton) {
    console.log("添加注册按钮点击事件");
    registerButton.onclick = function (e) {
      e.preventDefault(); // 阻止默认行为
      console.log("注册按钮被点击");
      handleRegister(); // 调用注册处理函数
      return false; // 防止表单提交
    };
  }

  // 处理注册表单提交 - 尝试另一种方式绑定事件
  if (registerFormElement) {
    registerFormElement.onsubmit = function (e) {
      e.preventDefault();
      console.log("注册表单提交");
      handleRegister();
      return false;
    };
  }

  // 注册处理函数
  async function handleRegister() {
    console.log("开始处理注册");
    registerMessage.textContent = "";
    registerMessage.className = "form-message";

    const username = document.getElementById("reg-username").value;
    const email = document.getElementById("reg-email").value;
    const password = document.getElementById("reg-password").value;
    const confirmPassword = document.getElementById(
      "reg-confirm-password"
    ).value;

    console.log("注册表单数据:", {
      username,
      email,
      password: "******",
      confirmPassword: "******",
    });

    // Validate form
    if (!username || !email || !password || !confirmPassword) {
      registerMessage.textContent = "请填写所有必填字段";
      return;
    }

    if (password !== confirmPassword) {
      registerMessage.textContent = "两次输入的密码不一致";
      return;
    }

    // Validate email format
    const emailRegex = /^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$/;
    if (!emailRegex.test(email)) {
      registerMessage.textContent = "请输入有效的邮箱地址";
      return;
    }

    // Validate password strength
    if (password.length < 6) {
      registerMessage.textContent = "密码长度至少为6个字符";
      return;
    }

    // Disable the form during submission
    setFormLoading(registerFormElement, true);

    try {
      // 添加日志以帮助调试
      console.log("发送注册请求:", { email, username, password: "******" });

      const response = await fetch("/api/v1/user/register", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({ email, username, password }),
      });

      console.log("注册请求已发送，状态码:", response.status);

      // 尝试读取响应，并处理可能的错误
      let data;
      try {
        const textResponse = await response.text();
        console.log("原始响应:", textResponse);
        data = JSON.parse(textResponse);
      } catch (parseError) {
        console.error("解析响应失败:", parseError);
        throw new Error("服务器响应格式错误");
      }

      console.log("注册响应:", data);

      if (data.code === 200) {
        registerMessage.textContent = "注册成功！正在自动登录...";
        registerMessage.className = "form-message success";

        // 自动登录
        setTimeout(async () => {
          try {
            const loginResponse = await fetch("/api/v1/user/login", {
              method: "POST",
              headers: {
                "Content-Type": "application/json",
              },
              body: JSON.stringify({ username, password }),
            });

            const loginData = await loginResponse.json();

            if (loginData.code === 200) {
              console.log("注册后自动登录成功");

              // Save JWT token
              JWTUtils.saveToken(loginData.data.token);

              // Redirect to chat page
              window.location.href = "/";
            } else {
              console.log("注册后自动登录失败:", loginData.msg);
              // 仍然跳转到登录页面
              loginTab.click();
              loginMessage.textContent = "注册成功，请登录";
              loginMessage.className = "form-message success";
              document.getElementById("email").value = username;
            }
          } catch (error) {
            console.error("自动登录错误:", error);
            // 跳转到登录页面
            loginTab.click();
            loginMessage.textContent = "注册成功，请登录";
            loginMessage.className = "form-message success";
            document.getElementById("email").value = username;
          }
        }, 1500);
      } else {
        if (
          data.msg &&
          (data.msg.includes("Email already registered") ||
            data.msg.includes("already registered"))
        ) {
          registerMessage.textContent =
            "此邮箱已被注册，请使用其他邮箱或直接登录";
        } else {
          registerMessage.textContent = data.msg || "注册失败，请稍后再试";
        }
      }
    } catch (error) {
      console.error("注册错误:", error);
      registerMessage.textContent =
        "注册时发生错误，请稍后再试: " + error.message;
    } finally {
      setFormLoading(registerFormElement, false);
    }
  }

  /**
   * 设置随机背景图片
   */
  function setRandomBackgroundImage() {
    // 背景图片列表
    const backgroundImages = [
      "../pictures/backgrounds/login/01.png",
      "../pictures/backgrounds/login/8e7a06969b6ecad17eef1914434859693493265644980448.png",
      "../pictures/backgrounds/login/dMIq1HzJ.png",
      "../pictures/backgrounds/login/X0_6rTZl.png",
    ];

    // 随机选择一个背景图片
    const randomIndex = Math.floor(Math.random() * backgroundImages.length);
    const selectedImage = backgroundImages[randomIndex];

    console.log("设置随机背景图片:", selectedImage);

    // 图片预加载和错误处理
    const img = new Image();
    img.onload = function () {
      // 图片加载成功，设置背景
      document.body.style.backgroundImage = `url(${selectedImage})`;
      document.body.style.backgroundSize = "cover";
      document.body.style.backgroundPosition = "center";
      document.body.style.backgroundRepeat = "no-repeat";
      document.body.style.backgroundAttachment = "fixed";
    };

    img.onerror = function () {
      console.error("背景图片加载失败:", selectedImage);
      // 如果图片加载失败，尝试使用备用图片或设置纯色背景
      document.body.style.backgroundColor = "#f0f2f5";
    };

    // 开始加载图片
    img.src = selectedImage;
  }

  /**
   * Set loading state for a form
   */
  function setFormLoading(form, isLoading) {
    const buttons = form.querySelectorAll("button");
    const inputs = form.querySelectorAll("input");

    buttons.forEach((button) => {
      button.disabled = isLoading;
      if (isLoading) {
        button.dataset.originalText = button.textContent;
        if (button.type === "submit") {
          button.textContent = "处理中...";
        }
      } else if (button.dataset.originalText) {
        button.textContent = button.dataset.originalText;
      }
    });

    inputs.forEach((input) => {
      input.disabled = isLoading;
    });
  }

  /**
   * Auto-fill login form if credentials are saved
   */
  function autoFillLoginForm() {
    const rememberedUsername = localStorage.getItem("remembered_username");
    const rememberedPassword = localStorage.getItem("remembered_password");
    const rememberedFlag = localStorage.getItem("remember_me");

    if (rememberedUsername && rememberedPassword && rememberedFlag === "true") {
      const usernameField = document.getElementById("email");
      const passwordField = document.getElementById("password");

      if (usernameField && passwordField && rememberMeCheckbox) {
        usernameField.value = rememberedUsername;
        try {
          passwordField.value = atob(rememberedPassword); // 解密
        } catch (e) {
          console.error("解密密码失败:", e);
          // 出错时清除存储的密码
          localStorage.removeItem("remembered_password");
        }
        rememberMeCheckbox.checked = true;

        // Special case for root user
        if (rememberedUsername === "root") {
          console.log("检测到root用户凭据，准备特殊处理");
        }
      }
    }
  }

  /**
   * Check if user is already logged in
   */
  function checkLoginStatus() {
    const currentUser = JWTUtils.getCurrentUser();

    if (currentUser) {
      console.log("检测到有效的登录状态:", currentUser);
      // User is already logged in, redirect to chat page
      window.location.href = "/";
    } else {
      // Try auto login with remembered credentials
      const rememberedUsername = localStorage.getItem("remembered_username");
      const rememberedPassword = localStorage.getItem("remembered_password");
      const rememberMe = localStorage.getItem("remember_me");

      if (rememberedUsername && rememberedPassword && rememberMe === "true") {
        // Perform auto login
        console.log("正在使用已保存的凭据自动登录...");

        // We'll attempt auto-login after the DOM is ready
        setTimeout(async () => {
          try {
            let password;
            try {
              password = atob(rememberedPassword);
            } catch (e) {
              console.error("解密密码失败:", e);
              return;
            }

            // Special handling for root user
            console.log("自动登录用户:", rememberedUsername);

            const response = await fetch("/api/v1/user/login", {
              method: "POST",
              headers: {
                "Content-Type": "application/json",
              },
              body: JSON.stringify({
                username: rememberedUsername,
                password: password,
              }),
            });

            const data = await response.json();

            if (data.code === 200) {
              console.log("自动登录成功");

              // Save JWT token
              JWTUtils.saveToken(data.data.token);

              // Redirect to chat page
              window.location.href = "/";
            } else {
              console.log("自动登录失败，需要手动登录:", data.msg);
              // 不要清除凭据，可能是临时服务器错误
            }
          } catch (error) {
            console.error("自动登录错误:", error);
          }
        }, 500);
      }
    }
  }

  // 处理修改密码表单提交
  if (resetPwdFormElement) {
    resetPwdFormElement.addEventListener("submit", async (e) => {
      e.preventDefault();
      resetPwdMessage.textContent = "";
      resetPwdMessage.className = "form-message";

      const username = document.getElementById("reset-email").value;
      const oldPassword = document.getElementById("reset-old-password").value;
      const newPassword = document.getElementById("reset-new-password").value;
      const confirmPassword = document.getElementById(
        "reset-confirm-password"
      ).value;

      // 表单验证
      if (!username || !oldPassword || !newPassword || !confirmPassword) {
        resetPwdMessage.textContent = "请填写所有必填字段";
        return;
      }

      if (newPassword !== confirmPassword) {
        resetPwdMessage.textContent = "两次输入的新密码不一致";
        return;
      }

      // 验证密码强度
      if (newPassword.length < 6) {
        resetPwdMessage.textContent = "新密码长度至少为6个字符";
        return;
      }

      // 禁用表单在提交期间
      setFormLoading(resetPwdFormElement, true);

              try {
          console.log("发送修改密码请求:", { username });

          // Get JWT token for authentication
          const token = JWTUtils.getStoredToken();
          if (!token) {
            resetPwdMessage.textContent = "请先登录";
            return;
          }

          const response = await fetch("/api/v1/user/password", {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
              "Authorization": token,
            },
            body: JSON.stringify({ old_password: oldPassword, new_password: newPassword }),
          });

        const data = await response.json();
        console.log("修改密码响应:", data);

        if (data.code === 200) {
          resetPwdMessage.textContent = "密码修改成功！请使用新密码登录";
          resetPwdMessage.className = "form-message success";

          // 如果有记住密码，更新存储的密码
          const currentUser = JWTUtils.getCurrentUser();
          if (currentUser && localStorage.getItem("remembered_username") === currentUser.username) {
            localStorage.setItem("remembered_password", btoa(newPassword));
          }

          // 清空表单并显示登录页
          setTimeout(() => {
            resetPwdFormElement.reset();
            resetPwdForm.style.display = "none";
            loginForm.style.display = "block";
            loginTab.classList.add("active");
            registerTab.classList.remove("active");
            loginMessage.textContent = "密码已更新，请使用新密码登录";
            loginMessage.className = "form-message success";
            document.getElementById("email").value = currentUser ? currentUser.username : username;
          }, 2000);
        } else {
          resetPwdMessage.textContent =
            data.msg || "密码修改失败，请检查您的邮箱和当前密码";
        }
      } catch (error) {
        console.error("修改密码错误:", error);
        resetPwdMessage.textContent = "修改密码时发生错误，请稍后再试";
      } finally {
        setFormLoading(resetPwdFormElement, false);
      }
    });
  }
});
