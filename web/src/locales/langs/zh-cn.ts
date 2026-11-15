const local: App.I18n.Schema = {
  system: {
    title: 'jzero-admin',
    updateTitle: '系统版本更新通知',
    updateContent: '检测到系统有新版本发布，是否立即刷新页面？',
    updateConfirm: '立即刷新',
    updateCancel: '稍后再说'
  },
  common: {
    action: '操作',
    add: '新增',
    addSuccess: '添加成功',
    backToHome: '返回首页',
    batchDelete: '批量删除',
    cancel: '取消',
    close: '关闭',
    check: '勾选',
    expandColumn: '展开列',
    columnSetting: '列设置',
    config: '配置',
    confirm: '确认',
    delete: '删除',
    deleteSuccess: '删除成功',
    confirmDelete: '确认删除吗？',
    download: '下载',
    edit: '编辑',
    editSuccess: '编辑成功',
    editHomeSuccess: '编辑角色首页成功',
    warning: '警告',
    error: '错误',
    index: '序号',
    keywordSearch: '请输入关键词搜索',
    logout: '退出登录',
    logoutConfirm: '确认退出登录吗？',
    lookForward: '敬请期待',
    modify: '修改',
    modifySuccess: '修改成功',
    noData: '无数据',
    operate: '操作',
    pleaseCheckValue: '请检查输入的值是否合法',
    refresh: '刷新',
    reset: '重置',
    search: '搜索',
    switch: '切换',
    tip: '提示',
    trigger: '触发',
    update: '更新',
    updateSuccess: '更新成功',
    userCenter: '个人中心',
    yesOrNo: {
      yes: '是',
      no: '否'
    }
  },
  request: {
    logout: '请求失败后登出用户',
    logoutMsg: '用户状态失效，请重新登录',
    logoutWithModal: '请求失败后弹出模态框再登出用户',
    logoutWithModalMsg: '用户状态失效，请重新登录',
    refreshToken: '请求的token已过期，刷新token',
    tokenExpired: 'token已过期'
  },
  theme: {
    themeSchema: {
      title: '主题模式',
      light: '亮色模式',
      dark: '暗黑模式',
      auto: '跟随系统'
    },
    grayscale: '灰色模式',
    colourWeakness: '色弱模式',
    layoutMode: {
      title: '布局模式',
      vertical: '左侧菜单模式',
      'vertical-mix': '左侧菜单混合模式',
      horizontal: '顶部菜单模式',
      'horizontal-mix': '顶部菜单混合模式',
      reverseHorizontalMix: '一级菜单与子级菜单位置反转'
    },
    recommendColor: '应用推荐算法的颜色',
    recommendColorDesc: '推荐颜色的算法参照',
    themeColor: {
      title: '主题颜色',
      primary: '主色',
      info: '信息色',
      success: '成功色',
      warning: '警告色',
      error: '错误色',
      followPrimary: '跟随主色'
    },
    scrollMode: {
      title: '滚动模式',
      wrapper: '外层滚动',
      content: '主体滚动'
    },
    page: {
      animate: '页面切换动画',
      mode: {
        title: '页面切换动画类型',
        'fade-slide': '滑动',
        fade: '淡入淡出',
        'fade-bottom': '底部消退',
        'fade-scale': '缩放消退',
        'zoom-fade': '渐变',
        'zoom-out': '闪现',
        none: '无'
      }
    },
    fixedHeaderAndTab: '固定头部和标签栏',
    header: {
      height: '头部高度',
      breadcrumb: {
        visible: '显示面包屑',
        showIcon: '显示面包屑图标'
      }
    },
    tab: {
      visible: '显示标签栏',
      cache: '缓存标签页',
      height: '标签栏高度',
      mode: {
        title: '标签栏风格',
        chrome: '谷歌风格',
        button: '按钮风格'
      }
    },
    sider: {
      inverted: '深色侧边栏',
      width: '侧边栏宽度',
      collapsedWidth: '侧边栏折叠宽度',
      mixWidth: '混合布局侧边栏宽度',
      mixCollapsedWidth: '混合布局侧边栏折叠宽度',
      mixChildMenuWidth: '混合布局子菜单宽度'
    },
    footer: {
      visible: '显示底部',
      fixed: '固定底部',
      height: '底部高度',
      right: '底部局右'
    },
    watermark: {
      visible: '显示全屏水印',
      text: '水印文本'
    },
    themeDrawerTitle: '主题配置',
    pageFunTitle: '页面功能',
    configOperation: {
      copyConfig: '复制配置',
      copySuccessMsg: '复制成功，请替换 src/theme/settings.ts 中的变量 themeSettings',
      resetConfig: '重置配置',
      resetSuccessMsg: '重置成功'
    }
  },
  route: {
    login: '登录',
    403: '无权限',
    404: '页面不存在',
    500: '服务器错误',
    'iframe-page': '外链页面',
    home: '首页',
    document: '文档',
    document_project: '项目文档',
    'document_project-link': '项目文档(外链)',
    document_vue: 'Vue文档',
    document_vite: 'Vite文档',
    document_unocss: 'UnoCSS文档',
    document_naive: 'Naive UI文档',
    document_antd: 'Ant Design Vue文档',
    'user-center': '个人中心',
    about: '关于',
    manage: '系统管理',
    manage_user: '用户管理',
    'manage_user-detail': '用户详情',
    manage_role: '角色管理',
    manage_menu: '菜单管理',
    hrm: '人力资源',
    hrm_org: '组织',
    hrm_org_orginfo: '组织管理',
    'hrm_org_org-chart': '组织架构图',
    hrm_attendance: '假勤管理',
    hrm_attendance_setting: '假勤设置',
    hrm_attendance_setting_group: '考勤组',
    'hrm_attendance_setting_attendance-config': '组织默认规则',
    hrm_attendance_setting_shift: '班次管理',
    'hrm_attendance_setting_leave-type': '假期类型',
    'hrm_attendance_setting_overtime-rule': '加班规则',
    'hrm_employee-relation': '人事管理',
    'hrm_employee-relation_transfer-list': '人事调动',
    'hrm_employee-relation_regular-list': '员工转正',
    hrm_performance: '绩效管理',
    hrm_performance_setting: '绩效设置',
    hrm_performance_setting_indicator: '指标库',
    hrm_performance_setting_template: '模板管理',
    hrm_performance_scheme: '绩效方案',
    hrm_training: '培训管理',
    hrm_training_my: '我的培训',
    hrm_reward: '奖惩管理',
    hrm_reward_bill: '奖惩单',
    hrm_recruitment: '招聘管理',
    'hrm_recruitment_requirement-list': '招聘需求',
    'hrm_recruitment_candidate-list': '候选人中心',
    hrm_salary: '薪酬管理',
    hrm_salary_setting: '薪酬设置',
    'hrm_salary_setting_subject': '薪酬主体',
    'hrm_salary_setting_item': '薪资项目',
    hrm_training_plan: '培训计划',
    hrm_training_session: '培训班次',
    'hrm_training_cert-template': '证书模板',
    'hrm_training_cert-issue': '证书发放',
    hrm_employee: '人事档案',
    'hrm_employee_employee-archive-list': '员工档案',
    exception: '异常页',
    exception_403: '403',
    exception_404: '404',
    exception_500: '500'
  },
  button: {
    manage: {
      user: {
        list: '用户列表',
        add: '新增用户',
        delete: '删除用户',
        edit: '编辑用户'
      },
      menu: {
        list: '菜单列表',
        add: '新增菜单',
        delete: '删除菜单',
        edit: '编辑菜单'
      },
      role: {
        list: '角色列表',
        add: '新增角色',
        delete: '删除角色',
        edit: '编辑角色'
      }
    },
    hrm: {
      org: {
        add: '新增组织',
        addchild: '新增下级组织',
        edit: '编辑组织',
        delete: '删除组织',
        import: '导入组织',
        export: '导出组织'
      },
      shift: {
        add: '新增班次',
        edit: '编辑班次',
        delete: '删除班次'
      },
      leavetype: {
        add: '新增假期类型',
        edit: '编辑假期类型',
        delete: '删除假期类型'
      },
      attendancegroup: {
        add: '新增考勤组',
        edit: '编辑考勤组',
        delete: '删除考勤组',
        togglestatus: '考勤组启停',
        members: '生效人数查询',
        shiftusage: '班次引用查询'
      },
      attendanceconfig: {
        save: '保存组织规则',
        delete: '删除组织规则'
      },
      overtimerule: {
        add: '新增加班规则',
        edit: '编辑加班规则',
        delete: '删除加班规则',
        togglestatus: '加班规则启停'
      },
      transfer: {
        add: '新增调动',
        edit: '编辑调动',
        delete: '删除调动'
      },
      regularization: {
        add: '新增转正单',
        edit: '编辑转正单',
        delete: '删除转正单'
      },
      reward: {
        add: '新增奖惩单',
        edit: '编辑奖惩单',
        delete: '删除奖惩单'
      },
      requirement: {
        add: '新增招聘需求',
        edit: '编辑招聘需求',
        delete: '删除招聘需求'
      },
      candidate: {
        add: '新增候选人',
        edit: '编辑候选人',
        delete: '删除候选人',
        import: '导入候选人',
        updatestatus: '候选人状态流转'
      },
      bill: {
        submit: '提交单据',
        approve: '审批通过',
        reject: '审批驳回',
        revoke: '撤销单据'
      },
      training: {
        downloadCert: '下载/预览证书扫描件',
        plan: {
          add: '新增计划',
          edit: '编辑计划',
          delete: '删除计划'
        },
        session: {
          add: '新增班次',
          edit: '编辑班次',
          delete: '删除班次',
          participants: '班次学员'
        },
        certTemplate: {
          add: '新增模板',
          edit: '编辑模板',
          toggleStatus: '启用/停用模板',
          delete: '删除模板'
        },
        certificate: {
          batchIssue: '批量发证',
          import: '导入外部证书',
          issue: '发放证书',
          revoke: '撤销证书',
          uploadFile: '上传证书扫描件'
        }
      },
      salary_subject: {
        add: '新增薪酬主体',
        edit: '编辑薪酬主体',
        delete: '删除薪酬主体'
      },
      salary_item: {
        add: '新增薪资项目',
        edit: '编辑薪资项目',
        delete: '删除薪资项目'
      },
      employee: {
        add: '新增员工',
        edit: '编辑员工',
        delete: '删除员工',
        import: '导入员工',
        export: '导出员工',
        generateuser: '生成用户'
      },
      performance: {
        indicator: {
          add: '新增指标',
          edit: '编辑指标',
          delete: '删除指标'
        },
        template: {
          add: '新增模板',
          edit: '编辑模板',
          delete: '删除模板'
        },
        scheme: {
          add: '新增方案',
          edit: '编辑方案',
          delete: '删除方案'
        }
      }
    }
  },
  page: {
    login: {
      common: {
        loginOrRegister: '登录 / 注册',
        usernamePlaceholder: '请输入用户名',
        phonePlaceholder: '请输入手机号',
        codePlaceholder: '请输入验证码',
        emailPlaceholder: '请输入邮箱',
        passwordPlaceholder: '请输入密码',
        confirmPasswordPlaceholder: '请再次输入密码',
        codeLogin: '验证码登录',
        confirm: '确定',
        back: '返回',
        validateSuccess: '验证成功',
        registerSuccess: '注册成功',
        loginSuccess: '登录成功',
        welcomeBack: '欢迎回来，{username} ！'
      },
      pwdLogin: {
        title: '密码登录',
        rememberMe: '记住我',
        forgetPassword: '忘记密码？',
        register: '注册账号',
        otherAccountLogin: '其他账号登录',
        otherLoginMode: '其他登录方式',
        superAdmin: '超级管理员',
        admin: '管理员',
        user: '普通用户'
      },
      codeLogin: {
        emailType: '邮箱',
        phoneType: '手机号',
        title: '{type}验证码登录',
        getCode: '获取验证码',
        reGetCode: '{time}秒后重新获取',
        sendCodeSuccess: '验证码发送成功',
        imageCodePlaceholder: '请输入图片验证码'
      },
      register: {
        title: '注册账号',
        agreement: '我已经仔细阅读并接受',
        protocol: '《用户协议》',
        policy: '《隐私权政策》'
      },
      resetPwd: {
        title: '重置密码'
      },
      bindWeChat: {
        title: '绑定微信'
      }
    },
    about: {
      title: '关于',
      introduction: `jzero-admin 基于 jzero 脚手架搭建后台系统并配备代码生成, 前端基于 soybean-admin 搭建。`,
      projectInfo: {
        title: '项目信息',
        version: '版本',
        latestBuildTime: '最新构建时间',
        githubLink: 'Github 地址',
        previewLink: '预览地址'
      },
      prdDep: '生产依赖',
      devDep: '开发依赖'
    },
    home: {
      branchDesc:
        '为了方便大家开发和更新合并，我们对main分支的代码进行了精简，只保留了首页菜单，其余内容已移至example分支进行维护。预览地址显示的内容即为example分支的内容。',
      greeting: '早安，{username}, 今天又是充满活力的一天!',
      weatherDesc: '今日多云转晴，20℃ - 25℃!',
      projectNews: {
        title: '项目里程碑',
        moreNews: '更多动态',
        desc1: '2024 年 04 月 07 日创建了开源项目 jzero!',
        desc2: '2024 年 10 月 14 号创建了开源项目 jzero-admin!',
        desc3: '2025 年不断的改进完善ing!',
        desc4: '2025 年 11 月 30 号创建了开源项目 jzero-intellij!',
        desc5: '2026 年 01 月 01 号发布了 jzero v1.0.0!'
      },
      creativity: '创意'
    },
    manage: {
      common: {
        status: {
          enable: '启用',
          disable: '禁用'
        }
      },
      role: {
        title: '角色列表',
        roleName: '角色名称',
        roleCode: '角色编码',
        roleStatus: '角色状态',
        roleDesc: '角色描述',
        menuAuth: '菜单权限',
        buttonAuth: '按钮权限',
        form: {
          roleName: '请输入角色名称',
          roleCode: '请输入角色编码',
          roleStatus: '请选择角色状态',
          roleDesc: '请输入角色描述'
        },
        addRole: '新增角色',
        editRole: '编辑角色'
      },
      user: {
        title: '用户列表',
        username: '用户名',
        userGender: '性别',
        nickName: '昵称',
        userPhone: '手机号',
        userEmail: '邮箱',
        userStatus: '用户状态',
        userRole: '用户角色',
        password: '密码',
        form: {
          username: '请输入用户名',
          userGender: '请选择性别',
          nickName: '请输入昵称',
          userPhone: '请输入手机号',
          userEmail: '请输入邮箱',
          userStatus: '请选择用户状态',
          userRole: '请选择用户角色',
          password: '请输入密码'
        },
        addUser: '新增用户',
        editUser: '编辑用户',
        gender: {
          male: '男',
          female: '女'
        }
      },
      menu: {
        home: '首页',
        title: '菜单列表',
        id: 'ID',
        parentId: '父级菜单ID',
        menuType: '菜单类型',
        menuName: '菜单名称',
        routeName: '路由名称',
        routePath: '路由路径',
        pathParam: '路径参数',
        layout: '布局',
        page: '页面组件',
        i18nKey: '国际化key',
        icon: '图标',
        localIcon: '本地图标',
        iconTypeTitle: '图标类型',
        order: '排序',
        constant: '常量路由',
        keepAlive: '缓存路由',
        href: '外链',
        iframePage: 'iframe page 链接',
        hideInMenu: '隐藏菜单',
        activeMenu: '高亮的菜单',
        multiTab: '支持多页签',
        fixedIndexInTab: '固定在页签中的序号',
        query: '路由参数',
        button: '按钮',
        buttonCode: '按钮编码',
        buttonDesc: '按钮描述',
        permission: '权限标识',
        permissionCode: '权限编码',
        permissionDesc: '权限描述',
        menuStatus: '菜单状态',
        form: {
          home: '请选择首页',
          menuType: '请选择菜单类型',
          menuName: '请输入菜单名称',
          routeName: '请输入路由名称',
          routePath: '请输入路由路径',
          pathParam: '请输入路径参数',
          page: '请选择页面组件',
          layout: '请选择布局组件',
          i18nKey: '请输入国际化key',
          icon: '请输入图标',
          localIcon: '请选择本地图标',
          order: '请输入排序',
          keepAlive: '请选择是否缓存路由',
          href: '请输入外链',
          hideInMenu: '请选择是否隐藏菜单',
          activeMenu: '请选择高亮的菜单的路由名称',
          multiTab: '请选择是否支持多标签',
          fixedInTab: '请选择是否固定在页签中',
          fixedIndexInTab: '请输入固定在页签中的序号',
          queryKey: '请输入路由参数Key',
          queryValue: '请输入路由参数Value',
          button: '请选择是否按钮',
          buttonCode: '请输入按钮编码',
          buttonDesc: '请输入按钮描述',
          permissionCode: '请输入权限编码',
          permissionDesc: '请输入权限描述',
          menuStatus: '请选择菜单状态'
        },
        addMenu: '新增菜单',
        editMenu: '编辑菜单',
        addChildMenu: '添加下一级',
        addButton: '新增按钮',
        type: {
          directory: '目录',
          menu: '菜单',
          button: '按钮'
        },
        iconType: {
          iconify: 'iconify图标',
          local: '本地图标'
        }
      }
    },
    hrm: {
      reward: {
        title: '奖惩单',
        billNo: '单据编号',
        billStatus: {
          title: '单据状态'
        },
        employeeNo: '员工工号',
        employeeName: '员工姓名',
        rewardType: {
          title: '奖惩类型',
          reward: '奖励',
          punishment: '处分'
        },
        rewardLevel: {
          title: '奖惩级别',
          company: '公司级',
          dept: '部门级',
          notice: '通报',
          other: '其他'
        },
        docNo: '文号',
        occurDate: '发生日期',
        amount: '金额',
        reason: '奖惩原因',
        salaryMonth: '入账薪资年月',
        synced: '薪资同步',
        creatorName: '申请人',
        createTime: '创建时间',
        yes: '是',
        no: '否',
        submit: '提交',
        approve: '通过',
        reject: '驳回',
        revoke: '撤销',
        records: '审批记录',
        noRecords: '暂无审批记录',
        confirmSubmit: '确认提交该奖惩单进入审批？',
        confirmApprove: '确认审批通过？通过后将同步为薪资项目的奖惩奖励/奖惩扣款数据源。',
        confirmReject: '确认驳回该奖惩单？',
        confirmRevoke: '确认撤销该奖惩单？',
        addReward: '新增奖惩单',
        editReward: '编辑奖惩单',
        sectionEmployee: '员工信息',
        sectionReward: '奖惩信息',
        form: {
          billNo: '请输入单据编号',
          status: '请选择单据状态',
          rewardType: '请选择奖惩类型',
          rewardLevel: '请选择奖惩级别',
          employeeNo: '请输入员工工号',
          employeeName: '请输入员工姓名',
          docNo: '请输入文号',
          salaryMonth: '入账薪资年月 YYYY-MM，留空默认为发生日期所在月',
          reason: '请输入奖惩原因'
        }
      },
      requirement: {
        title: '招聘需求',
        billNo: '需求编号',
        billStatus: {
          title: '单据状态'
        },
        postName: '岗位名称',
        postCategory: {
          title: '岗位类别',
          tech: '技术类',
          product: '产品类',
          design: '设计类',
          operation: '运营类',
          market: '市场类',
          function: '职能类',
          other: '其他'
        },
        headcount: '需求人数',
        urgency: {
          title: '紧急度',
          normal: '普通',
          urgent: '紧急'
        },
        expectDate: '期望到岗日期',
        orgName: '所属部门',
        requirementDesc: '需求说明',
        creatorName: '申请人',
        createTime: '创建时间',
        submit: '提交',
        approve: '通过',
        reject: '驳回',
        revoke: '撤销',
        records: '审批记录',
        noRecords: '暂无审批记录',
        confirmSubmit: '确认提交该招聘需求单进入审批？',
        confirmApprove: '确认审批通过？通过后该需求可被招聘职位引用发布。',
        confirmReject: '确认驳回该招聘需求单？',
        confirmRevoke: '确认撤销该招聘需求单？',
        addRequirement: '新增招聘需求',
        editRequirement: '编辑招聘需求',
        sectionPost: '岗位信息',
        form: {
          billNo: '请输入需求编号',
          status: '请选择单据状态',
          postName: '请输入岗位名称',
          postCategory: '请选择岗位类别',
          urgency: '请选择紧急度',
          org: '请选择所属部门',
          requirementDesc: '请输入需求说明/任职要求'
        }
      },
      candidate: {
        title: '候选人中心',
        name: '姓名',
        phone: '手机号',
        email: '邮箱',
        sourceChannel: {
          title: '来源渠道',
          recommend: '内部推荐',
          website: '招聘网站',
          campus: '校园招聘',
          headhunter: '猎头',
          other: '其他'
        },
        education: {
          title: '最高学历',
          doctor: '博士',
          master: '硕士',
          bachelor: '本科',
          associate: '大专',
          senior: '高中',
          other: '其他'
        },
        currentCompany: '当前公司',
        currentPosition: '当前岗位',
        expectSalary: '期望薪资',
        remark: '备注',
        status: {
          title: '状态',
          new: '新线索',
          delivered: '已投递',
          interviewing: '面试中',
          offer: '已发Offer',
          hired: '录用通过',
          rejected: '不通过'
        },
        createTime: '创建时间',
        advanceTo: '推进到{status}',
        confirmAdvance: '确认将该候选人的状态推进到 {status}？',
        confirmReject: '确认标记该候选人不通过？',
        reject: '不通过',
        import: 'Excel导入',
        importModalTitle: '导入候选人',
        importSuccess: '导入成功',
        importTemplateTip:
          '模板列：姓名、手机号、邮箱、来源渠道（内部推荐/招聘网站/校园招聘/猎头/其他）、最高学历（博士/硕士/本科/大专/高中/其他）、当前公司、当前岗位、期望薪资',
        chooseFileTip: '请选择 xlsx 文件',
        addCandidate: '新增候选人',
        editCandidate: '编辑候选人',
        form: {
          name: '请输入姓名',
          phone: '请输入手机号',
          email: '请输入邮箱',
          sourceChannel: '请选择来源渠道',
          education: '请选择最高学历',
          currentCompany: '请输入当前公司',
          currentPosition: '请输入当前岗位',
          expectSalary: '请输入期望薪资（如 20k-30k）',
          remark: '请输入备注'
        }
      },
      performance: {
        common: {
          status: {
            enable: '启用',
            disable: '停用'
          }
        },
        indicatorType: {
          kpi: '结果指标(KPI)',
          behavior: '行为能力'
        },
        scoreRule: {
          formula: '公式评分',
          manual: '手工评分'
        },
        dataSource: {
          system: '系统取数',
          manual: '手工填报'
        },
        schemePeriodType: {
          month: '月度',
          quarter: '季度',
          half_year: '半年',
          year: '年度'
        },
        schemeAssessmentMode: {
          kpi: 'KPI',
          kpi_behavior: 'KPI+行为混合'
        },
        schemeScoreMode: {
          percent: '百分制',
          grade: '等级制',
          amount: '金额制'
        },
        schemePayLinkMode: {
          coefficient: '系数制',
          amount: '金额制'
        },
        indicator: {
          title: '指标库',
          code: '指标编码',
          name: '指标名称',
          indicatorType: '指标类型',
          unit: '计量单位',
          scoreRule: '评分规则',
          dataSource: '数据来源',
          defaultWeight: '默认权重',
          status: '状态',
          remark: '备注',
          form: {
            code: '请输入指标编码',
            name: '请输入指标名称',
            indicatorType: '请选择指标类型',
            unit: '请输入计量单位',
            scoreRule: '请选择评分规则',
            dataSource: '请选择数据来源',
            status: '请选择状态',
            remark: '请输入备注'
          },
          addIndicator: '新增指标',
          editIndicator: '编辑指标'
        },
        template: {
          title: '考核模板',
          name: '模板名称',
          status: '状态',
          remark: '备注',
          createTime: '创建时间',
          form: {
            name: '请输入模板名称',
            status: '请选择状态',
            remark: '请输入备注'
          },
          configItems: '指标权重配置',
          indicatorPlaceholder: '请选择指标',
          addIndicatorItem: '添加指标',
          weightSumLabel: '权重之和',
          weightSumError: '模板内指标权重之和必须等于 100%',
          itemIncomplete: '请为每条指标配置选择指标并填写大于 0 的权重',
          addTemplate: '新增模板',
          editTemplate: '编辑模板'
        },
        scheme: {
          title: '绩效方案',
          name: '方案名称',
          templateName: '关联模板',
          periodType: '周期类型',
          assessmentMode: '评估模式',
          scoreMode: '评分模式',
          payLinkMode: '联薪方式',
          status: '状态',
          remark: '备注',
          form: {
            name: '请输入方案名称',
            templateUuid: '请选择关联模板',
            periodType: '请选择周期类型',
            assessmentMode: '请选择评估模式',
            scoreMode: '请选择评分模式',
            payLinkMode: '请选择联薪方式',
            status: '请选择状态',
            remark: '请输入备注'
          },
          addScheme: '新增方案',
          editScheme: '编辑方案'
        }
      },
      org: {
        title: '组织管理',
        orgName: '组织名称',
        orgTypeTitle: '组织类型',
        orgType: {
          company: '公司',
          department: '部门'
        },
        leader: '负责人',
        orderNum: '显示顺序',
        orgStatus: '组织状态',
        status: {
          open: '开启',
          closed: '关闭'
        },
        remark: '备注',
        createTime: '创建时间',
        parentOrg: '上级组织',
        form: {
          orgName: '请输入组织名称',
          orgType: '请选择组织类型',
          leader: '请输入负责人',
          orderNum: '请输入显示顺序',
          orgStatus: '请选择组织状态',
          remark: '请输入备注',
          parentOrg: '上级组织'
        },
        addOrg: '新增组织',
        addChildOrg: '新增下级',
        editOrg: '编辑组织',
        importOrg: '导入',
        exportOrg: '导出',
        collapse: '收缩',
        expand: '展开',
        importSuccess: '导入成功',
        chooseFileTip: '请选择 xlsx 文件',
        importModalTitle: '导入组织',
        orgChart: {
          title: '组织架构图',
          memberCount: '本级人数',
          memberCountTip: '直属本组织的在册员工数',
          totalMemberCount: '汇总人数',
          totalMemberCountTip: '本级及全部下级组织的在册员工数',
          expandAll: '展开全部',
          collapseAll: '全部收起',
          refresh: '刷新',
          empty: '暂无组织数据，请先在组织管理中维护组织',
          noLeader: '未设置'
        }
      },
      employee: {
        title: '员工档案',
        employeeNo: '员工工号',
        name: '姓名',
        genderLabel: '性别',
        gender: {
          male: '男',
          female: '女'
        },
        orgName: '所属部门',
        companyName: '所属单位',
        position: '职位',
        post: '职务',
        personStatusLabel: '人员状态',
        personStatus: {
          regular: '正式',
          probation: '试用期',
          intern: '实习生',
          temp: '临时工',
          resigned: '离职',
          retired: '退休'
        },
        hireDate: '入职日期',
        formalDate: '转正日期',
        phone: '手机号',
        email: '邮箱',
        userStatusLabel: '用户状态',
        userStatus: {
          created: '已生成',
          none: '未生成'
        },
        birthday: '生日',
        nation: '民族',
        politicalStatus: '政治面貌',
        maritalStatusLabel: '婚姻状况',
        maritalStatus: {
          unmarried: '未婚',
          married: '已婚',
          divorced: '离异',
          widowed: '丧偶'
        },
        nativePlace: '籍贯',
        bloodType: '血型',
        idCard: '身份证号',
        householdAddress: '户口地址',
        address: '现住址',
        emergencyContact: '紧急联系人',
        emergencyPhone: '紧急联系人电话',
        bankName: '工资开户行',
        bankAccount: '工资卡账户',
        avatar: '头像',
        remark: '备注',
        createTime: '创建时间',
        addEmployee: '新增员工',
        editEmployee: '编辑员工',
        importEmployee: '导入员工',
        exportEmployee: '导出',
        generateUser: '生成用户',
        generateUserSelectTip: '请先勾选要生成账号的员工',
        generateUserSuccess: '已生成 {generated} 个系统账号, 跳过 {skipped} 个已有账号的员工',
        generateUserAllSkipped: '所选员工均已生成过账号 (跳过 {skipped} 个)',
        importSuccess: '导入成功',
        chooseFileTip: '请选择 xlsx 文件',
        importModalTitle: '导入员工',
        tabBasic: '基本信息',
        tabEducation: '教育经历',
        tabWork: '工作经历',
        tabFamily: '家属',
        addEducationRow: '添加教育经历',
        noEducationRows: '暂无教育经历',
        addWorkRow: '添加工作经历',
        noWorkRows: '暂无工作经历',
        addFamilyRow: '添加家属',
        noFamilyRows: '暂无家属信息',
        eduSchool: '学校',
        eduMajor: '专业',
        eduEducation: '学历',
        eduStartDate: '开始日期',
        eduEndDate: '结束日期',
        workCompany: '单位',
        workPosition: '职位',
        workStartDate: '开始日期',
        workEndDate: '结束日期',
        familyName: '姓名',
        familyRelation: '关系',
        familyPhone: '电话',
        familyWorkUnit: '工作单位',
        form: {
          employeeNo: '工号（留空自动生成）',
          employeeNoAuto: '留空则按最大工号+1自动生成',
          name: '请输入姓名',
          org: '请选择所属部门',
          company: '请选择所属单位',
          position: '请输入职位',
          post: '请输入职务',
          personStatus: '请选择人员状态',
          phone: '请输入手机号',
          email: '请输入邮箱',
          birthday: '请输入生日（如 1990-01-01）',
          nation: '请输入民族',
          politicalStatus: '请输入政治面貌',
          maritalStatus: '请选择婚姻状况',
          nativePlace: '请输入籍贯',
          bloodType: '请输入血型',
          idCard: '请输入身份证号',
          householdAddress: '请输入户口地址',
          address: '请输入现住址',
          emergencyContact: '请输入紧急联系人',
          emergencyPhone: '请输入紧急联系人电话',
          bankName: '请输入工资开户行',
          bankAccount: '请输入工资卡账户',
          remark: '请输入备注'
        }
      },
      attendance: {
        yes: '是',
        no: '否',
        status: {
          enabled: '启用',
          disabled: '停用'
        },
        group: {
          title: '考勤组',
          groupName: '考勤组名称',
          groupType: {
            title: '班制',
            fixed: '固定班制',
            scheduling: '排班制',
            flexible: '自由工时'
          },
          workTime: '考勤时间',
          workDays: '考勤日',
          scope: '适用范围',
          scopeDepts: '部门×{n}',
          scopeMembers: '人员×{n}',
          managers: '负责人',
          status: '状态',
          remark: '备注',
          enable: '启用',
          disable: '停用',
          confirmEnable: '确认启用该考勤组？',
          confirmDisable: '停用后考勤回退到组织默认规则，确认停用？',
          flexibleNoTime: '自由工时（不打卡）',
          schedulingFallback: '按排班表执行',
          members: '生效人数',
          noMembers: '暂无生效人员',
          addGroup: '新增考勤组',
          editGroup: '编辑考勤组',
          weekday: {
            1: '周一',
            2: '周二',
            3: '周三',
            4: '周四',
            5: '周五',
            6: '周六',
            7: '周日'
          },
          weekdayShort: {
            1: '一',
            2: '二',
            3: '三',
            4: '四',
            5: '五',
            6: '六',
            7: '日'
          },
          form: {
            groupName: '请输入考勤组名称',
            groupType: '请选择班制',
            workDays: '请选择考勤日',
            orgScope: '适用部门（含下级）',
            memberScope: '指定人员',
            managers: '负责人（回车添加）',
            status: '请选择状态',
            remark: '请输入备注'
          }
        },
        config: {
          title: '组织默认规则',
          org: '组织',
          orgName: '组织名称',
          quickStart: '快速开始',
          editConfig: '编辑组织默认规则',
          recheck: '重新体检',
          fallbackHint:
            '规则优先级：考勤组 > 组织默认规则；未加入考勤组的员工按本部门规则执行；本部门没设置则用最近上级组织的规则。',
          noRule: '无规则',
          sourceGroup: '考勤组',
          sourceOrg: '组织规则',
          healthEmployee: '员工',
          healthSource: '规则来源',
          healthyTotal: '共 {n} 人',
          healthyGroup: '考勤组命中 {n}',
          healthyOrg: '组织规则 {n}',
          healthyNoRule: '无规则 {n}',
          form: {
            orgName: '请输入组织名称',
            org: '请选择组织'
          }
        },
        shift: {
          title: '班次管理',
          shiftName: '班次名称',
          workTime: '上下班时间',
          crossDay: '跨天',
          nextDay: '次日',
          restTime: '休息时段',
          noRest: '无休息',
          standardHours: '标准工时',
          flexibleMinutes: '弹性',
          notFlexible: '不弹性',
          flexibleMinutesValue: '弹性 {n} 分钟',
          status: '状态',
          remark: '备注',
          addShift: '新增班次',
          editShift: '编辑班次',
          form: {
            shiftName: '请输入班次名称',
            startTime: '请选择上班时间',
            endTime: '请选择下班时间',
            rest: '休息时段（可选）',
            flexibleMinutes: '请输入弹性分钟数',
            status: '请选择状态',
            remark: '请输入备注'
          }
        },
        leaveType: {
          title: '假期类型',
          typeName: '类型名称',
          typeCode: '类型编码',
          typeValue: '类型值',
          balanceEnabled: '余额管理',
          attachmentRequired: '需要附件',
          defaultDays: '默认天数',
          orderNum: '排序',
          status: '状态',
          remark: '备注',
          addLeaveType: '新增假期类型',
          editLeaveType: '编辑假期类型',
          form: {
            typeName: '请输入类型名称',
            typeCode: '请输入类型编码（如 annual_leave）',
            typeValue: '请输入类型值',
            status: '请选择状态',
            remark: '请输入备注'
          }
        },
        overtimeRule: {
          title: '加班规则',
          ruleName: '规则名称',
          workday: '工作日',
          restday: '休息日',
          holiday: '法定节假日',
          dayPolicy: '当天策略',
          approvalRequired: '审批方式',
          punchRequired: '打卡要求',
          compMultiplier: '调休倍数',
          multiplierTip: '调休×N = 加班 1 小时折算 N 小时调休',
          multiplierValue: '调休×{n}',
          needApproval: '需审批',
          noApproval: '免审批',
          needPunch: '需打卡',
          noPunch: '免打卡',
          minUnitMinutes: '最小单位',
          minUnitHalfHour: '30分钟',
          minUnitHour: '1小时',
          monthlyLimitHours: '月上限',
          monthlyLimitTip: '0 表示不限制',
          noLimit: '不限制',
          usedGroups: '使用的考勤组',
          boundGroupCount: '{n} 个考勤组',
          globalFallback: '全部考勤组',
          addRule: '新增加班规则',
          editRule: '编辑加班规则',
          enable: '启用',
          disable: '停用',
          status: '状态',
          remark: '备注',
          form: {
            ruleName: '请输入规则名称',
            status: '请选择状态',
            remark: '请输入备注'
          }
        }
      },
      employeeRelation: {
        billStatus: {
          title: '单据状态',
          unsubmitted: '未提交',
          approving: '审批中',
          approved: '审批通过',
          rejected: '审批不通过',
          revoked: '已撤销'
        },
        transfer: {
          title: '人事调动',
          billNo: '单据编号',
          employeeNo: '员工工号',
          employeeName: '员工姓名',
          changeType: {
            title: '异动类型',
            post: '调岗',
            org: '调部门',
            salary: '调薪',
            company: '调公司',
            other: '其他'
          },
          originalCompany: '原公司',
          originalOrg: '原部门',
          originalPosition: '原职位',
          originalPost: '原岗位',
          newCompany: '变更为公司',
          newOrg: '变更为部门',
          newPosition: '变更为职位',
          newPost: '变更为岗位',
          newSalary: '调整后薪资',
          transferReason: '调动原因',
          effectiveImmediately: '是否立即生效',
          effectiveDate: '生效日期',
          applied: '已生效',
          creatorName: '申请人',
          createTime: '创建时间',
          yes: '是',
          no: '否',
          submit: '提交',
          approve: '通过',
          reject: '驳回',
          revoke: '撤销',
          records: '审批记录',
          noRecords: '暂无审批记录',
          confirmSubmit: '确认提交该调动单进入审批？',
          confirmApprove: '确认审批通过？立即生效的调动单将同步更新员工档案。',
          confirmReject: '确认驳回该调动单？',
          confirmRevoke: '确认撤销该调动单？',
          addTransfer: '新增调动',
          editTransfer: '编辑调动',
          sectionEmployee: '员工信息',
          sectionOriginal: '原（变动前）',
          sectionNew: '新（变动后）',
          sectionEffect: '生效设置',
          form: {
            billNo: '请输入单据编号',
            status: '请选择单据状态',
            employeeNo: '请输入员工工号',
            employeeName: '请输入员工姓名',
            changeType: '请选择异动类型',
            org: '请选择组织',
            transferReason: '请输入调动原因',
            effectiveDateRequired: '非立即生效时必须选择生效日期'
          }
        },
        regularization: {
          title: '员工转正',
          billNo: '单据编号',
          employeeNo: '员工工号',
          employeeName: '员工姓名',
          selectEmployee: '选择员工',
          orgUuid: '员工所属部门',
          companyUuid: '员工所属公司',
          position: '职位',
          post: '职务',
          personStatus: '人员状态',
          hireDate: '入职日期',
          formalDate: '转正日期',
          expectedFormalDate: '预计转正日期',
          remark: '备注',
          creatorName: '申请人',
          createTime: '创建时间',
          submit: '提交',
          approve: '通过',
          reject: '驳回',
          revoke: '撤销',
          records: '审批记录',
          noRecords: '暂无审批记录',
          confirmSubmit: '确认提交该转正单进入审批？',
          confirmApprove: '确认审批通过？通过后将把员工人员状态更新为正式并写入转正日期。',
          confirmReject: '确认驳回该转正单？',
          confirmRevoke: '确认撤销该转正单？',
          addRegularization: '新增转正单',
          editRegularization: '编辑转正单',
          sectionEmployee: '员工信息',
          sectionSnapshot: '档案快照（自动带出）',
          sectionRegular: '转正信息',
          form: {
            billNo: '请输入单据编号',
            status: '请选择单据状态',
            employeeNo: '请输入员工工号',
            selectEmployee: '请选择员工',
            remark: '请输入备注',
            none: '选择员工后自动带出'
          }
        }
      },
      salary: {
        subject: {
          title: '薪酬主体',
          subjectName: '主体名称',
          orgName: '关联公司',
          payAccount: '发薪账户',
          payBank: '发薪银行',
          status: '状态',
          remark: '备注',
          addSubject: '新增薪酬主体',
          editSubject: '编辑薪酬主体',
          form: {
            subjectName: '请输入主体名称',
            orgName: '请选择关联公司',
            payAccount: '请输入发薪账户',
            payBank: '请输入发薪银行',
            status: '请选择状态',
            remark: '请输入备注'
          }
        },
        item: {
          title: '薪资项目',
          itemName: '项目名称',
          itemCode: '项目编码',
          itemType: {
            title: '项目类型',
            fixed: '固定项',
            data: '取数项',
            formula: '公式项'
          },
          itemClass: {
            title: '项目分类',
            earning: '应发',
            deduction: '应扣',
            cost: '成本'
          },
          valueMode: {
            title: '取值方式',
            manual: '手工维护',
            system: '系统取数',
            formula: '公式计算'
          },
          calcOrder: '计算顺序',
          orderNum: '显示顺序',
          formula: '公式表达式',
          isPreset: '预置',
          presetYes: '预置',
          presetNo: '自定义',
          status: '状态',
          remark: '备注',
          addItem: '新增薪资项目',
          editItem: '编辑薪资项目',
          form: {
            itemName: '请输入项目名称',
            itemCode: '请输入项目编码（如 OTHER_BONUS）',
            itemType: '请选择项目类型',
            itemClass: '请选择项目分类',
            valueMode: '请选择取值方式',
            formula: '请输入公式表达式',
            status: '请选择状态',
            remark: '请输入备注'
          }
        }
      },
      training: {
        mySessionTitle: '我的班次',
        myCertTitle: '我的培训证书',
        sessionNo: '班次编号',
        sessionName: '班次',
        location: '地点',
        lecturer: '讲师',
        hours: '学时',
        startTime: '开始时间',
        endTime: '结束时间',
        sessionStatusLabel: '状态',
        sessionStatus: {
          notStarted: '未开始',
          ongoing: '进行中',
          finished: '已结束'
        },
        certName: '证书名称',
        certNo: '证书号',
        issueTime: '发证时间',
        certStatusLabel: '状态',
        certStatus: {
          pending: '待发',
          issued: '已发',
          revoked: '已撤销'
        },
        scanFile: '扫描件',
        preview: '预览',
        noFile: '无扫描件',
        planType: {
          special: '专项',
          newHire: '新员工培养',
          safety: '安全培训',
          skill: '技能提升'
        },
        published: {
          draft: '草稿',
          published: '已发布'
        },
        plan: {
          title: '培训计划',
          billNo: '单据编号',
          billStatusLabel: '单据状态',
          planName: '计划名称',
          theme: '主题',
          planYear: '年度',
          planType: '计划类型',
          startDate: '开始日期',
          endDate: '结束日期',
          published: '发布状态',
          creatorName: '申请人',
          addPlan: '新增计划',
          editPlan: '编辑计划',
          submit: '提交',
          approve: '审批通过',
          reject: '审批驳回',
          revoke: '撤销',
          records: '审批记录',
          confirmSubmit: '确认提交该培训计划进入审批？',
          confirmApprove: '确认审批通过？通过后计划将发布（草稿→已发布）。',
          confirmReject: '确认审批驳回？',
          confirmRevoke: '确认撤销该单据？',
          form: {
            planName: '请输入计划名称',
            theme: '请输入主题',
            planYear: '请输入年度（如 2026）',
            planType: '请选择计划类型',
            status: '请选择单据状态',
            published: '请选择发布状态',
            endDateBeforeStart: '结束日期不能早于开始日期'
          }
        },
        session: {
          title: '培训班次',
          sessionName: '班次名称',
          planName: '培训计划',
          lecturer: '讲师',
          remark: '备注',
          addSession: '新增班次',
          editSession: '编辑班次',
          participants: '选人',
          finish: '结束',
          confirmFinish: '确认结束该班次？结束后不可恢复。',
          participantsTip: '保存将全量覆盖班次学员（选人结果即最终名单）。',
          form: {
            sessionName: '请输入班次名称',
            planUuid: '请选择培训计划（仅已发布）',
            lecturer: '请输入讲师',
            location: '请输入地点',
            status: '请选择状态',
            participantsPlaceholder: '搜索并选择学员（账号/姓名）',
            endTimeBeforeStart: '结束时间必须晚于开始时间'
          }
        },
        certTemplate: {
          title: '证书模板',
          templateName: '模板名称',
          certType: '证书类型',
          sourceLabel: '来源',
          noPrefix: '编号前缀',
          statusLabel: '状态',
          remark: '备注',
          addTemplate: '新增模板',
          editTemplate: '编辑模板',
          enable: '启用',
          disable: '停用',
          confirmEnable: '确认启用该模板？启用后可用于发证。',
          confirmDisable: '确认停用该模板？停用后不可发证。',
          source: {
            system: '系统',
            org: '企业'
          },
          status: {
            enabled: '启用',
            disabled: '停用'
          },
          form: {
            templateName: '请输入模板名称',
            certType: '请输入证书类型',
            source: '请选择来源',
            noPrefix: '请输入编号前缀（如 JN/AQ/HF，证书号=前缀-年-流水）'
          }
        },
        certificate: {
          title: '证书发放',
          student: '学员',
          template: '证书模板',
          sourceLabel: '来源',
          source: {
            session: '班次结业',
            import: '外部导入'
          },
          session: '班次',
          batchIssue: '批量发证',
          batchIssueDone: '本次发证',
          batchIssueSkipped: '跳过（已有证书）',
          batchIssueTip: '按所选班次的结业学员批量生成证书（证书号=模板前缀-年-流水）；已持有该模板有效证书的学员自动跳过，可重复调用（幂等）。',
          importExternal: '导入外部证书',
          issue: '发放',
          revoke: '撤销',
          upload: '上传扫描件',
          chooseFile: '选择文件',
          confirmIssue: '确认发放该证书？发放后可在我的培训中查看。',
          confirmRevoke: '确认撤销该证书？撤销后学员端显示已撤销（终态）。',
          form: {
            session: '请选择班次（已结束）',
            template: '请选择证书模板（启用中）',
            student: '搜索并选择学员（账号/姓名）',
            certName: '请输入证书名称',
            certType: '请输入证书类型',
            certNo: '请输入外部证书号（唯一）',
            status: '请选择状态',
            source: '请选择来源'
          }
        }
      }
    }
  },
  form: {
    required: '不能为空',
    username: {
      required: '请输入用户名',
      invalid: '用户名格式不正确'
    },
    phone: {
      required: '请输入手机号',
      invalid: '手机号格式不正确'
    },
    pwd: {
      required: '请输入密码',
      invalid: '密码格式不正确，6-18位字符，包含字母、数字、下划线'
    },
    confirmPwd: {
      required: '请输入确认密码',
      invalid: '两次输入密码不一致'
    },
    code: {
      required: '请输入验证码',
      invalid: '验证码格式不正确'
    },
    email: {
      required: '请输入邮箱',
      invalid: '邮箱格式不正确'
    }
  },
  dropdown: {
    closeCurrent: '关闭',
    closeOther: '关闭其它',
    closeLeft: '关闭左侧',
    closeRight: '关闭右侧',
    closeAll: '关闭所有'
  },
  icon: {
    themeConfig: '主题配置',
    themeSchema: '主题模式',
    lang: '切换语言',
    fullscreen: '全屏',
    fullscreenExit: '退出全屏',
    reload: '刷新页面',
    collapse: '折叠菜单',
    expand: '展开菜单',
    pin: '固定',
    unpin: '取消固定'
  },
  datatable: {
    itemCount: '共 {total} 条'
  }
};

export default local;
