const local: App.I18n.Schema = {
  system: {
    title: 'jzero-admin',
    updateTitle: 'System Version Update Notification',
    updateContent: 'A new version of the system has been detected. Do you want to refresh the page immediately?',
    updateConfirm: 'Refresh immediately',
    updateCancel: 'Later'
  },
  common: {
    action: 'Action',
    add: 'Add',
    addSuccess: 'Add Success',
    backToHome: 'Back to home',
    batchDelete: 'Batch Delete',
    cancel: 'Cancel',
    close: 'Close',
    check: 'Check',
    expandColumn: 'Expand Column',
    columnSetting: 'Column Setting',
    config: 'Config',
    confirm: 'Confirm',
    delete: 'Delete',
    deleteSuccess: 'Delete Success',
    confirmDelete: 'Are you sure you want to delete?',
    download: 'Download',
    edit: 'Edit',
    editSuccess: 'Edit Success',
    editHomeSuccess: 'Edit Role Home Success',
    warning: 'Warning',
    error: 'Error',
    index: 'Index',
    keywordSearch: 'Please enter keyword',
    logout: 'Logout',
    logoutConfirm: 'Are you sure you want to log out?',
    lookForward: 'Coming soon',
    modify: 'Modify',
    modifySuccess: 'Modify Success',
    noData: 'No Data',
    operate: 'Operate',
    pleaseCheckValue: 'Please check whether the value is valid',
    refresh: 'Refresh',
    reset: 'Reset',
    search: 'Search',
    switch: 'Switch',
    tip: 'Tip',
    trigger: 'Trigger',
    update: 'Update',
    updateSuccess: 'Update Success',
    userCenter: 'User Center',
    yesOrNo: {
      yes: 'Yes',
      no: 'No'
    }
  },
  request: {
    logout: 'Logout user after request failed',
    logoutMsg: 'User status is invalid, please log in again',
    logoutWithModal: 'Pop up modal after request failed and then log out user',
    logoutWithModalMsg: 'User status is invalid, please log in again',
    refreshToken: 'The requested token has expired, refresh the token',
    tokenExpired: 'The requested token has expired'
  },
  theme: {
    themeSchema: {
      title: 'Theme Schema',
      light: 'Light',
      dark: 'Dark',
      auto: 'Follow System'
    },
    grayscale: 'Grayscale',
    colourWeakness: 'Colour Weakness',
    layoutMode: {
      title: 'Layout Mode',
      vertical: 'Vertical Menu Mode',
      horizontal: 'Horizontal Menu Mode',
      'vertical-mix': 'Vertical Mix Menu Mode',
      'horizontal-mix': 'Horizontal Mix menu Mode',
      reverseHorizontalMix: 'Reverse first level menus and child level menus position'
    },
    recommendColor: 'Apply Recommended Color Algorithm',
    recommendColorDesc: 'The recommended color algorithm refers to',
    themeColor: {
      title: 'Theme Color',
      primary: 'Primary',
      info: 'Info',
      success: 'Success',
      warning: 'Warning',
      error: 'Error',
      followPrimary: 'Follow Primary'
    },
    scrollMode: {
      title: 'Scroll Mode',
      wrapper: 'Wrapper',
      content: 'Content'
    },
    page: {
      animate: 'Page Animate',
      mode: {
        title: 'Page Animate Mode',
        fade: 'Fade',
        'fade-slide': 'Slide',
        'fade-bottom': 'Fade Zoom',
        'fade-scale': 'Fade Scale',
        'zoom-fade': 'Zoom Fade',
        'zoom-out': 'Zoom Out',
        none: 'None'
      }
    },
    fixedHeaderAndTab: 'Fixed Header And Tab',
    header: {
      height: 'Header Height',
      breadcrumb: {
        visible: 'Breadcrumb Visible',
        showIcon: 'Breadcrumb Icon Visible'
      }
    },
    tab: {
      visible: 'Tab Visible',
      cache: 'Tab Cache',
      height: 'Tab Height',
      mode: {
        title: 'Tab Mode',
        chrome: 'Chrome',
        button: 'Button'
      }
    },
    sider: {
      inverted: 'Dark Sider',
      width: 'Sider Width',
      collapsedWidth: 'Sider Collapsed Width',
      mixWidth: 'Mix Sider Width',
      mixCollapsedWidth: 'Mix Sider Collapse Width',
      mixChildMenuWidth: 'Mix Child Menu Width'
    },
    footer: {
      visible: 'Footer Visible',
      fixed: 'Fixed Footer',
      height: 'Footer Height',
      right: 'Right Footer'
    },
    watermark: {
      visible: 'Watermark Full Screen Visible',
      text: 'Watermark Text'
    },
    themeDrawerTitle: 'Theme Configuration',
    pageFunTitle: 'Page Function',
    configOperation: {
      copyConfig: 'Copy Config',
      copySuccessMsg: 'Copy Success, Please replace the variable "themeSettings" in "src/theme/settings.ts"',
      resetConfig: 'Reset Config',
      resetSuccessMsg: 'Reset Success'
    }
  },
  route: {
    login: 'Login',
    403: 'No Permission',
    404: 'Page Not Found',
    500: 'Server Error',
    'iframe-page': 'Iframe',
    home: 'Home',
    document: 'Document',
    document_project: 'Project Document',
    'document_project-link': 'Project Document(External Link)',
    document_vue: 'Vue Document',
    document_vite: 'Vite Document',
    document_unocss: 'UnoCSS Document',
    document_naive: 'Naive UI Document',
    document_antd: 'Ant Design Vue Document',
    'user-center': 'User Center',
    about: 'About',
    manage: 'System Manage',
    manage_user: 'User Manage',
    'manage_user-detail': 'User Detail',
    manage_role: 'Role Manage',
    manage_menu: 'Menu Manage',
    hrm: 'HRM',
    hrm_org: 'Organization',
    hrm_org_orginfo: 'Org Management',
    hrm_attendance: 'Attendance',
    hrm_attendance_setting: 'Attendance Settings',
    hrm_attendance_setting_shift: 'Shifts',
    'hrm_attendance_setting_leave-type': 'Leave Types',
    hrm_performance: 'Performance',
    hrm_performance_setting: 'Settings',
    hrm_performance_setting_indicator: 'Indicator Library',
    hrm_performance_setting_template: 'Template Management',
    hrm_training: 'Training',
    hrm_training_my: 'My Training',
    exception: 'Exception',
    exception_403: '403',
    exception_404: '404',
    exception_500: '500'
  },
  button: {
    manage: {
      user: {
        list: 'User List',
        add: 'Add User',
        edit: 'Edit User',
        delete: 'Delete User'
      },
      menu: {
        list: 'Menu List',
        add: 'Add Menu',
        edit: 'Edit Menu',
        delete: 'Delete Menu'
      },
      role: {
        list: 'Role List',
        add: 'Add Role',
        edit: 'Edit Role',
        delete: 'Delete Role'
      }
    },
    hrm: {
      org: {
        add: 'Add Org',
        addchild: 'Add Child Org',
        edit: 'Edit Org',
        delete: 'Delete Org',
        import: 'Import Orgs',
        export: 'Export Orgs'
      },
      shift: {
        add: 'Add Shift',
        edit: 'Edit Shift',
        delete: 'Delete Shift'
      },
      leavetype: {
        add: 'Add Leave Type',
        edit: 'Edit Leave Type',
        delete: 'Delete Leave Type'
      },
      training: {
        downloadCert: 'Download/Preview Cert Scan'
      },
      performance: {
        indicator: {
          add: 'Add Indicator',
          edit: 'Edit Indicator',
          delete: 'Delete Indicator'
        },
        template: {
          add: 'Add Template',
          edit: 'Edit Template',
          delete: 'Delete Template'
        }
      }
    }
  },
  page: {
    login: {
      common: {
        loginOrRegister: 'Login / Register',
        usernamePlaceholder: 'Please enter user name',
        phonePlaceholder: 'Please enter phone number',
        codePlaceholder: 'Please enter verification code',
        emailPlaceholder: 'Please enter email',
        passwordPlaceholder: 'Please enter password',
        confirmPasswordPlaceholder: 'Please enter password again',
        codeLogin: 'Verification code login',
        confirm: 'Confirm',
        back: 'Back',
        registerSuccess: 'Register success',
        validateSuccess: 'Verification passed',
        loginSuccess: 'Login successfully',
        welcomeBack: 'Welcome back, {username} !'
      },
      pwdLogin: {
        title: 'Password Login',
        rememberMe: 'Remember me',
        forgetPassword: 'Forget password?',
        register: 'Register',
        otherAccountLogin: 'Other Account Login',
        otherLoginMode: 'Other Login Mode',
        superAdmin: 'Super Admin',
        admin: 'Admin',
        user: 'User'
      },
      codeLogin: {
        title: 'Verification {type} Code Login',
        emailType: 'Email',
        phoneType: 'Phone',
        getCode: 'Get verification code',
        reGetCode: 'Reacquire after {time}s',
        sendCodeSuccess: 'Verification code sent successfully',
        imageCodePlaceholder: 'Please enter image verification code'
      },
      register: {
        title: 'Register',
        agreement: 'I have read and agree to',
        protocol: '《User Agreement》',
        policy: '《Privacy Policy》'
      },
      resetPwd: {
        title: 'Reset Password'
      },
      bindWeChat: {
        title: 'Bind WeChat'
      }
    },
    about: {
      title: 'About',
      introduction: `jzero-admin builds the backend system based on jzero scaffolding and is equipped with code generation, and the front end is built based on soybean-admin.`,
      projectInfo: {
        title: 'Project Info',
        version: 'Version',
        latestBuildTime: 'Latest Build Time',
        githubLink: 'Github Link',
        previewLink: 'Preview Link'
      },
      prdDep: 'Production Dependency',
      devDep: 'Development Dependency'
    },
    home: {
      branchDesc:
        'For the convenience of everyone in developing and updating the merge, we have streamlined the code of the main branch, only retaining the homepage menu, and the rest of the content has been moved to the example branch for maintenance. The preview address displays the content of the example branch.',
      greeting: 'Good morning, {username}, today is another day full of vitality!',
      weatherDesc: 'Today is cloudy to clear, 20℃ - 25℃!',
      projectNews: {
        title: 'Project News',
        moreNews: 'More News',
        desc1: 'Created the open source project jzero on April 07, 2024!',
        desc2: 'Created the open source project jzero-admin on October 14, 2024!',
        desc3: 'Continuously improving and perfecting in 2025!',
        desc4: 'Created the open source project jzero-intellij on November 30, 2025!',
        desc5: 'Released jzero v1.0.0 on January 01, 2026!'
      },
      creativity: 'Creativity'
    },
    manage: {
      common: {
        status: {
          enable: 'Enable',
          disable: 'Disable'
        }
      },
      role: {
        title: 'Role List',
        roleName: 'Role Name',
        roleCode: 'Role Code',
        roleStatus: 'Role Status',
        roleDesc: 'Role Description',
        menuAuth: 'Menu Auth',
        buttonAuth: 'Button Auth',
        form: {
          roleName: 'Please enter role name',
          roleCode: 'Please enter role code',
          roleStatus: 'Please select role status',
          roleDesc: 'Please enter role description'
        },
        addRole: 'Add Role',
        editRole: 'Edit Role'
      },
      user: {
        title: 'User List',
        username: 'User Name',
        userGender: 'Gender',
        nickName: 'Nick Name',
        userPhone: 'Phone Number',
        userEmail: 'Email',
        userStatus: 'User Status',
        userRole: 'User Role',
        password: 'Password',
        form: {
          username: 'Please enter user name',
          userGender: 'Please select gender',
          nickName: 'Please enter nick name',
          userPhone: 'Please enter phone number',
          userEmail: 'Please enter email',
          userStatus: 'Please select user status',
          userRole: 'Please select user role',
          password: 'Please enter password'
        },
        addUser: 'Add User',
        editUser: 'Edit User',
        gender: {
          male: 'Male',
          female: 'Female'
        }
      },
      menu: {
        home: 'Home',
        title: 'Menu List',
        id: 'ID',
        parentId: 'Parent ID',
        menuType: 'Menu Type',
        menuName: 'Menu Name',
        routeName: 'Route Name',
        routePath: 'Route Path',
        pathParam: 'Path Param',
        layout: 'Layout Component',
        page: 'Page Component',
        i18nKey: 'I18n Key',
        icon: 'Icon',
        localIcon: 'Local Icon',
        iconTypeTitle: 'Icon Type',
        order: 'Order',
        constant: 'Constant',
        keepAlive: 'Keep Alive',
        href: 'Href',
        iframePage: 'Iframe Page Href',
        hideInMenu: 'Hide In Menu',
        activeMenu: 'Active Menu',
        multiTab: 'Multi Tab',
        fixedIndexInTab: 'Fixed Index In Tab',
        query: 'Query Params',
        button: 'Button',
        buttonCode: 'Button Code',
        buttonDesc: 'Button Desc',
        permission: 'permission',
        permissionCode: 'Permission Code',
        permissionDesc: 'Permission Desc',
        menuStatus: 'Menu Status',
        form: {
          home: 'Please select home',
          menuType: 'Please select menu type',
          menuName: 'Please enter menu name',
          routeName: 'Please enter route name',
          routePath: 'Please enter route path',
          pathParam: 'Please enter path param',
          page: 'Please select page component',
          layout: 'Please select layout component',
          i18nKey: 'Please enter i18n key',
          icon: 'Please enter iconify name',
          localIcon: 'Please enter local icon name',
          order: 'Please enter order',
          keepAlive: 'Please select whether to cache route',
          href: 'Please enter href',
          hideInMenu: 'Please select whether to hide menu',
          activeMenu: 'Please select route name of the highlighted menu',
          multiTab: 'Please select whether to support multiple tabs',
          fixedInTab: 'Please select whether to fix in the tab',
          fixedIndexInTab: 'Please enter the index fixed in the tab',
          queryKey: 'Please enter route parameter Key',
          queryValue: 'Please enter route parameter Value',
          button: 'Please select whether it is a button',
          buttonCode: 'Please enter button code',
          buttonDesc: 'Please enter button description',
          permissionCode: 'Please enter permission code',
          permissionDesc: 'Please enter permission description',
          menuStatus: 'Please select menu status'
        },
        addMenu: 'Add Menu',
        editMenu: 'Edit Menu',
        addChildMenu: 'Add Child Menu',
        addButton: 'Add Button',
        type: {
          directory: 'Directory',
          menu: 'Menu',
          button: 'Button'
        },
        iconType: {
          iconify: 'Iconify Icon',
          local: 'Local Icon'
        }
      }
    },
    hrm: {
      performance: {
        common: {
          status: {
            enable: 'Enabled',
            disable: 'Disabled'
          }
        },
        indicatorType: {
          kpi: 'KPI',
          behavior: 'Behavior Competency'
        },
        scoreRule: {
          formula: 'Formula Scoring',
          manual: 'Manual Scoring'
        },
        dataSource: {
          system: 'System Data',
          manual: 'Manual Input'
        },
        indicator: {
          title: 'Indicator Library',
          code: 'Indicator Code',
          name: 'Indicator Name',
          indicatorType: 'Indicator Type',
          unit: 'Unit',
          scoreRule: 'Scoring Rule',
          dataSource: 'Data Source',
          defaultWeight: 'Default Weight',
          status: 'Status',
          remark: 'Remark',
          form: {
            code: 'Please enter indicator code',
            name: 'Please enter indicator name',
            indicatorType: 'Please select indicator type',
            unit: 'Please enter unit',
            scoreRule: 'Please select scoring rule',
            dataSource: 'Please select data source',
            status: 'Please select status',
            remark: 'Please enter remark'
          },
          addIndicator: 'Add Indicator',
          editIndicator: 'Edit Indicator'
        },
        template: {
          title: 'Assessment Template',
          name: 'Template Name',
          status: 'Status',
          remark: 'Remark',
          createTime: 'Create Time',
          form: {
            name: 'Please enter template name',
            status: 'Please select status',
            remark: 'Please enter remark'
          },
          configItems: 'Indicator Weight Configuration',
          indicatorPlaceholder: 'Please select indicator',
          addIndicatorItem: 'Add Indicator',
          weightSumLabel: 'Weight Sum',
          weightSumError: 'The sum of indicator weights in a template must equal 100%',
          itemIncomplete: 'Please select an indicator and enter a weight greater than 0 for each row',
          addTemplate: 'Add Template',
          editTemplate: 'Edit Template'
        }
      },
      org: {
        title: 'Org Management',
        orgName: 'Org Name',
        orgTypeTitle: 'Org Type',
        orgType: {
          company: 'Company',
          department: 'Department'
        },
        leader: 'Leader',
        orderNum: 'Display Order',
        orgStatus: 'Org Status',
        status: {
          open: 'Open',
          closed: 'Closed'
        },
        remark: 'Remark',
        createTime: 'Create Time',
        parentOrg: 'Parent Org',
        form: {
          orgName: 'Please enter org name',
          orgType: 'Please select org type',
          leader: 'Please enter leader',
          orderNum: 'Please enter display order',
          orgStatus: 'Please select org status',
          remark: 'Please enter remark',
          parentOrg: 'Parent Org'
        },
        addOrg: 'Add Org',
        addChildOrg: 'Add Child Org',
        editOrg: 'Edit Org',
        importOrg: 'Import',
        exportOrg: 'Export',
        collapse: 'Collapse',
        expand: 'Expand',
        importSuccess: 'Import succeeded',
        chooseFileTip: 'Please select an xlsx file',
        importModalTitle: 'Import Orgs'
      },
      attendance: {
        yes: 'Yes',
        no: 'No',
        status: {
          enabled: 'Enabled',
          disabled: 'Disabled'
        },
        shift: {
          title: 'Shifts',
          shiftName: 'Shift Name',
          workTime: 'Work Time',
          crossDay: 'Cross-day',
          nextDay: 'Next day',
          restTime: 'Rest Period',
          noRest: 'No rest',
          standardHours: 'Standard Hours',
          flexibleMinutes: 'Flexible',
          notFlexible: 'Not flexible',
          flexibleMinutesValue: '{n} min flexible',
          status: 'Status',
          remark: 'Remark',
          addShift: 'Add Shift',
          editShift: 'Edit Shift',
          form: {
            shiftName: 'Please enter shift name',
            startTime: 'Please select start time',
            endTime: 'Please select end time',
            rest: 'Rest period (optional)',
            flexibleMinutes: 'Please enter flexible minutes',
            status: 'Please select status',
            remark: 'Please enter remark'
          }
        },
        leaveType: {
          title: 'Leave Types',
          typeName: 'Type Name',
          typeCode: 'Type Code',
          typeValue: 'Type Value',
          balanceEnabled: 'Balance Management',
          attachmentRequired: 'Attachment Required',
          defaultDays: 'Default Days',
          orderNum: 'Order',
          status: 'Status',
          remark: 'Remark',
          addLeaveType: 'Add Leave Type',
          editLeaveType: 'Edit Leave Type',
          form: {
            typeName: 'Please enter type name',
            typeCode: 'Please enter type code (e.g. annual_leave)',
            typeValue: 'Please enter type value',
            status: 'Please select status',
            remark: 'Please enter remark'
          }
        }
      },
      training: {
        mySessionTitle: 'My Sessions',
        myCertTitle: 'My Certificates',
        sessionNo: 'Session No.',
        sessionName: 'Session',
        location: 'Location',
        lecturer: 'Lecturer',
        hours: 'Hours',
        startTime: 'Start Time',
        endTime: 'End Time',
        sessionStatusLabel: 'Status',
        sessionStatus: {
          notStarted: 'Not Started',
          ongoing: 'Ongoing',
          finished: 'Finished'
        },
        certName: 'Certificate Name',
        certNo: 'Certificate No.',
        issueTime: 'Issue Time',
        certStatusLabel: 'Status',
        certStatus: {
          pending: 'Pending',
          issued: 'Issued',
          revoked: 'Revoked'
        },
        scanFile: 'Scan File',
        preview: 'Preview',
        noFile: 'No scan file'
      }
    }
  },
  form: {
    required: 'Cannot be empty',
    username: {
      required: 'Please enter user name',
      invalid: 'User name format is incorrect'
    },
    phone: {
      required: 'Please enter phone number',
      invalid: 'Phone number format is incorrect'
    },
    pwd: {
      required: 'Please enter password',
      invalid: '6-18 characters, including letters, numbers, and underscores'
    },
    confirmPwd: {
      required: 'Please enter password again',
      invalid: 'The two passwords are inconsistent'
    },
    code: {
      required: 'Please enter verification code',
      invalid: 'Verification code format is incorrect'
    },
    email: {
      required: 'Please enter email',
      invalid: 'Email format is incorrect'
    }
  },
  dropdown: {
    closeCurrent: 'Close Current',
    closeOther: 'Close Other',
    closeLeft: 'Close Left',
    closeRight: 'Close Right',
    closeAll: 'Close All'
  },
  icon: {
    themeConfig: 'Theme Configuration',
    themeSchema: 'Theme Schema',
    lang: 'Switch Language',
    fullscreen: 'Fullscreen',
    fullscreenExit: 'Exit Fullscreen',
    reload: 'Reload Page',
    collapse: 'Collapse Menu',
    expand: 'Expand Menu',
    pin: 'Pin',
    unpin: 'Unpin'
  },
  datatable: {
    itemCount: 'Total {total} items'
  }
};

export default local;
