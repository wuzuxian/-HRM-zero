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
    hrm_attendance_setting_shift: '班次管理',
    'hrm_attendance_setting_leave-type': '假期类型',
    'hrm_employee-relation': '人事管理',
    'hrm_employee-relation_transfer-list': '人事调动',
    hrm_performance: '绩效管理',
    hrm_performance_setting: '绩效设置',
    hrm_performance_setting_indicator: '指标库',
    hrm_performance_setting_template: '模板管理',
    hrm_performance_scheme: '绩效方案',
    hrm_training: '培训管理',
    hrm_training_my: '我的培训',
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
      transfer: {
        add: '新增调动',
        edit: '编辑调动',
        delete: '删除调动'
      },
      bill: {
        submit: '提交单据',
        approve: '审批通过',
        reject: '审批驳回',
        revoke: '撤销单据'
      },
      training: {
        downloadCert: '下载/预览证书扫描件'
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
        noFile: '无扫描件'
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
