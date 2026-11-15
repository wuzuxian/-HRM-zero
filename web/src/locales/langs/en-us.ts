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
    'hrm_org_org-chart': 'Org Chart',
    hrm_attendance: 'Attendance',
    hrm_attendance_setting: 'Attendance Settings',
    hrm_attendance_setting_group: 'Attendance Groups',
    'hrm_attendance_setting_attendance-config': 'Org Default Rules',
    hrm_attendance_setting_shift: 'Shifts',
    'hrm_attendance_setting_leave-type': 'Leave Types',
    'hrm_attendance_setting_overtime-rule': 'Overtime Rules',
    hrm_attendance_scheduling: 'Scheduling',
    'hrm_attendance_scheduling_schedule-cycle': 'Schedule Cycles',
    'hrm_employee-relation': 'HR Management',
    'hrm_employee-relation_transfer-list': 'Transfer',
    'hrm_employee-relation_entry-list': 'Entry',
    hrm_performance: 'Performance',
    hrm_performance_setting: 'Settings',
    hrm_performance_setting_indicator: 'Indicator Library',
    hrm_performance_setting_template: 'Template Management',
    hrm_performance_scheme: 'Scheme',
    hrm_training: 'Training',
    hrm_training_my: 'My Training',
    hrm_reward: 'Reward & Discipline',
    hrm_reward_bill: 'Reward Bill',
    hrm_recruitment: 'Recruitment',
    'hrm_recruitment_requirement-list': 'Requirement',
    'hrm_recruitment_candidate-list': 'Candidates',
    hrm_salary: 'Salary',
    hrm_salary_setting: 'Salary Settings',
    hrm_salary_setting_subject: 'Salary Subjects',
    hrm_salary_setting_item: 'Salary Items',
    hrm_training_plan: 'Training Plan',
    hrm_training_session: 'Training Session',
    'hrm_training_cert-template': 'Certificate Template',
    'hrm_training_cert-issue': 'Certificate Issue',
    hrm_employee: 'HR Archive',
    'hrm_employee_employee-archive-list': 'Employee Archive',
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
      attendancegroup: {
        add: 'Add Attendance Group',
        edit: 'Edit Attendance Group',
        delete: 'Delete Attendance Group',
        togglestatus: 'Toggle Attendance Group',
        members: 'View Effective Members',
        shiftusage: 'View Shift Usage'
      },
      attendanceconfig: {
        save: 'Save Org Rule',
        delete: 'Delete Org Rule'
      },
      overtimerule: {
        add: 'Add Overtime Rule',
        edit: 'Edit Overtime Rule',
        delete: 'Delete Overtime Rule',
        togglestatus: 'Toggle Overtime Rule'
      },
      schedulecycle: {
        add: 'Add Schedule Cycle',
        edit: 'Edit Schedule Cycle',
        delete: 'Delete Schedule Cycle',
        togglestatus: 'Toggle Schedule Cycle'
      },
      transfer: {
        add: 'Add Transfer',
        edit: 'Edit Transfer',
        delete: 'Delete Transfer'
      },
      entry: {
        add: 'Add Entry Bill',
        edit: 'Edit Entry Bill',
        delete: 'Delete Entry Bill'
      },
      reward: {
        add: 'Add Reward Bill',
        edit: 'Edit Reward Bill',
        delete: 'Delete Reward Bill'
      },
      requirement: {
        add: 'Add Requirement',
        edit: 'Edit Requirement',
        delete: 'Delete Requirement'
      },
      candidate: {
        add: 'Add Candidate',
        edit: 'Edit Candidate',
        delete: 'Delete Candidate',
        import: 'Import Candidates',
        updatestatus: 'Candidate Status Transition'
      },
      bill: {
        submit: 'Submit Bill',
        approve: 'Approve',
        reject: 'Reject',
        revoke: 'Revoke'
      },
      training: {
        downloadCert: 'Download/Preview Cert Scan',
        plan: {
          add: 'Add Plan',
          edit: 'Edit Plan',
          delete: 'Delete Plan'
        },
        session: {
          add: 'Add Session',
          edit: 'Edit Session',
          delete: 'Delete Session',
          participants: 'Participants'
        },
        certTemplate: {
          add: 'Add Template',
          edit: 'Edit Template',
          toggleStatus: 'Enable/Disable Template',
          delete: 'Delete Template'
        },
        certificate: {
          batchIssue: 'Batch Issue',
          import: 'Import External Certificate',
          issue: 'Issue Certificate',
          revoke: 'Revoke Certificate',
          uploadFile: 'Upload Scan File'
        }
      },
      salary_subject: {
        add: 'Add Salary Subject',
        edit: 'Edit Salary Subject',
        delete: 'Delete Salary Subject'
      },
      salary_item: {
        add: 'Add Salary Item',
        edit: 'Edit Salary Item',
        delete: 'Delete Salary Item'
      },
      employee: {
        add: 'Add Employee',
        edit: 'Edit Employee',
        delete: 'Delete Employee',
        import: 'Import Employees',
        export: 'Export Employees',
        generateuser: 'Generate User'
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
        },
        scheme: {
          add: 'Add Scheme',
          edit: 'Edit Scheme',
          delete: 'Delete Scheme'
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
      reward: {
        title: 'Reward Bill',
        billNo: 'Bill No.',
        billStatus: {
          title: 'Bill Status'
        },
        employeeNo: 'Employee No.',
        employeeName: 'Employee Name',
        rewardType: {
          title: 'Type',
          reward: 'Reward',
          punishment: 'Discipline'
        },
        rewardLevel: {
          title: 'Level',
          company: 'Company Level',
          dept: 'Department Level',
          notice: 'Notice',
          other: 'Other'
        },
        docNo: 'Document No.',
        occurDate: 'Occur Date',
        amount: 'Amount',
        reason: 'Reason',
        salaryMonth: 'Salary Month',
        synced: 'Salary Synced',
        creatorName: 'Applicant',
        createTime: 'Created At',
        yes: 'Yes',
        no: 'No',
        submit: 'Submit',
        approve: 'Approve',
        reject: 'Reject',
        revoke: 'Revoke',
        records: 'Records',
        noRecords: 'No approval records',
        confirmSubmit: 'Submit this reward bill for approval?',
        confirmApprove: 'Approve this bill? Once approved it becomes the 奖惩奖励/奖惩扣款 salary-item data source.',
        confirmReject: 'Reject this reward bill?',
        confirmRevoke: 'Revoke this reward bill?',
        addReward: 'Add Reward Bill',
        editReward: 'Edit Reward Bill',
        sectionEmployee: 'Employee Info',
        sectionReward: 'Reward Info',
        form: {
          billNo: 'Please enter bill no.',
          status: 'Please select bill status',
          rewardType: 'Please select reward type',
          rewardLevel: 'Please select reward level',
          employeeNo: 'Please enter employee no.',
          employeeName: 'Please enter employee name',
          docNo: 'Please enter document no.',
          salaryMonth: 'Salary month YYYY-MM, defaults to the occur-date month',
          reason: 'Please enter reason'
        }
      },
      requirement: {
        title: 'Recruitment Requirement',
        billNo: 'Requirement No.',
        billStatus: {
          title: 'Bill Status'
        },
        postName: 'Post Name',
        postCategory: {
          title: 'Post Category',
          tech: 'Tech',
          product: 'Product',
          design: 'Design',
          operation: 'Operation',
          market: 'Marketing',
          function: 'Function',
          other: 'Other'
        },
        headcount: 'Headcount',
        urgency: {
          title: 'Urgency',
          normal: 'Normal',
          urgent: 'Urgent'
        },
        expectDate: 'Expected Onboard Date',
        orgName: 'Department',
        requirementDesc: 'Requirement Description',
        creatorName: 'Applicant',
        createTime: 'Create Time',
        submit: 'Submit',
        approve: 'Approve',
        reject: 'Reject',
        revoke: 'Revoke',
        records: 'Approval Records',
        noRecords: 'No approval records',
        confirmSubmit: 'Submit this requirement bill for approval?',
        confirmApprove: 'Approve this bill? Once approved it can be referenced by recruitment positions.',
        confirmReject: 'Reject this requirement bill?',
        confirmRevoke: 'Revoke this requirement bill?',
        addRequirement: 'Add Requirement',
        editRequirement: 'Edit Requirement',
        sectionPost: 'Post Info',
        form: {
          billNo: 'Please enter requirement no.',
          status: 'Please select bill status',
          postName: 'Please enter post name',
          postCategory: 'Please select post category',
          urgency: 'Please select urgency',
          org: 'Please select department',
          requirementDesc: 'Please enter requirement description'
        }
      },
      candidate: {
        title: 'Candidate Center',
        name: 'Name',
        phone: 'Phone',
        email: 'Email',
        sourceChannel: {
          title: 'Source Channel',
          recommend: 'Referral',
          website: 'Job Site',
          campus: 'Campus',
          headhunter: 'Headhunter',
          other: 'Other'
        },
        education: {
          title: 'Highest Education',
          doctor: 'Doctor',
          master: 'Master',
          bachelor: 'Bachelor',
          associate: 'Associate',
          senior: 'Senior High',
          other: 'Other'
        },
        currentCompany: 'Current Company',
        currentPosition: 'Current Position',
        expectSalary: 'Expected Salary',
        remark: 'Remark',
        status: {
          title: 'Status',
          new: 'New Lead',
          delivered: 'Delivered',
          interviewing: 'Interviewing',
          offer: 'Offer Sent',
          hired: 'Hired',
          rejected: 'Rejected'
        },
        createTime: 'Create Time',
        advanceTo: 'Advance to {status}',
        confirmAdvance: "Advance this candidate's status to {status}?",
        confirmReject: 'Mark this candidate as rejected?',
        reject: 'Reject',
        import: 'Excel Import',
        importModalTitle: 'Import Candidates',
        importSuccess: 'Import succeeded',
        importTemplateTip:
          'Template columns: Name, Phone, Email, Source Channel (Referral/Job Site/Campus/Headhunter/Other), Highest Education (Doctor/Master/Bachelor/Associate/Senior High/Other), Current Company, Current Position, Expected Salary',
        chooseFileTip: 'Please choose an xlsx file',
        addCandidate: 'Add Candidate',
        editCandidate: 'Edit Candidate',
        form: {
          name: 'Please enter name',
          phone: 'Please enter phone',
          email: 'Please enter email',
          sourceChannel: 'Please select source channel',
          education: 'Please select highest education',
          currentCompany: 'Please enter current company',
          currentPosition: 'Please enter current position',
          expectSalary: 'Please enter expected salary (e.g. 20k-30k)',
          remark: 'Please enter remark'
        }
      },
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
        schemePeriodType: {
          month: 'Monthly',
          quarter: 'Quarterly',
          half_year: 'Semi-annual',
          year: 'Annual'
        },
        schemeAssessmentMode: {
          kpi: 'KPI',
          kpi_behavior: 'KPI + Behavior'
        },
        schemeScoreMode: {
          percent: 'Percentage',
          grade: 'Grade',
          amount: 'Amount'
        },
        schemePayLinkMode: {
          coefficient: 'Coefficient',
          amount: 'Amount'
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
        },
        scheme: {
          title: 'Performance Scheme',
          name: 'Scheme Name',
          templateName: 'Linked Template',
          periodType: 'Period Type',
          assessmentMode: 'Assessment Mode',
          scoreMode: 'Score Mode',
          payLinkMode: 'Pay Link Mode',
          status: 'Status',
          remark: 'Remark',
          form: {
            name: 'Please enter scheme name',
            templateUuid: 'Please select a template',
            periodType: 'Please select period type',
            assessmentMode: 'Please select assessment mode',
            scoreMode: 'Please select score mode',
            payLinkMode: 'Please select pay link mode',
            status: 'Please select status',
            remark: 'Please enter remark'
          },
          addScheme: 'Add Scheme',
          editScheme: 'Edit Scheme'
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
        importModalTitle: 'Import Orgs',
        orgChart: {
          title: 'Org Chart',
          memberCount: 'Direct',
          memberCountTip: 'Employees directly under this org',
          totalMemberCount: 'Total',
          totalMemberCountTip: 'Employees of this org and all sub-orgs',
          expandAll: 'Expand all',
          collapseAll: 'Collapse all',
          refresh: 'Refresh',
          empty: 'No org data yet. Please maintain orgs in Org Management first.',
          noLeader: 'Not set'
        }
      },
      employee: {
        title: 'Employee Archive',
        employeeNo: 'Employee No.',
        name: 'Name',
        genderLabel: 'Gender',
        gender: {
          male: 'Male',
          female: 'Female'
        },
        orgName: 'Department',
        companyName: 'Company',
        position: 'Position',
        post: 'Job Post',
        personStatusLabel: 'Person Status',
        personStatus: {
          regular: 'Regular',
          probation: 'Probation',
          intern: 'Intern',
          temp: 'Temp',
          resigned: 'Resigned',
          retired: 'Retired'
        },
        hireDate: 'Hire Date',
        formalDate: 'Formal Date',
        phone: 'Phone',
        email: 'Email',
        userStatusLabel: 'User Status',
        userStatus: {
          created: 'Created',
          none: 'Not Created'
        },
        birthday: 'Birthday',
        nation: 'Nation',
        politicalStatus: 'Political Status',
        maritalStatusLabel: 'Marital Status',
        maritalStatus: {
          unmarried: 'Unmarried',
          married: 'Married',
          divorced: 'Divorced',
          widowed: 'Widowed'
        },
        nativePlace: 'Native Place',
        bloodType: 'Blood Type',
        idCard: 'ID Card No.',
        householdAddress: 'Household Address',
        address: 'Address',
        emergencyContact: 'Emergency Contact',
        emergencyPhone: 'Emergency Phone',
        bankName: 'Salary Bank',
        bankAccount: 'Salary Account',
        avatar: 'Avatar',
        remark: 'Remark',
        createTime: 'Created At',
        addEmployee: 'Add Employee',
        editEmployee: 'Edit Employee',
        importEmployee: 'Import',
        exportEmployee: 'Export',
        generateUser: 'Generate User',
        generateUserSelectTip: 'Select employees to generate accounts first',
        generateUserSuccess: 'Generated {generated} account(s), skipped {skipped} employee(s) with existing accounts',
        generateUserAllSkipped: 'All selected employees already have accounts (skipped {skipped})',
        importSuccess: 'Import succeeded',
        chooseFileTip: 'Please select an xlsx file',
        importModalTitle: 'Import Employees',
        tabBasic: 'Basic Info',
        tabEducation: 'Education',
        tabWork: 'Work Experience',
        tabFamily: 'Family',
        addEducationRow: 'Add education row',
        noEducationRows: 'No education rows',
        addWorkRow: 'Add work row',
        noWorkRows: 'No work rows',
        addFamilyRow: 'Add family row',
        noFamilyRows: 'No family rows',
        eduSchool: 'School',
        eduMajor: 'Major',
        eduEducation: 'Degree',
        eduStartDate: 'Start Date',
        eduEndDate: 'End Date',
        workCompany: 'Company',
        workPosition: 'Position',
        workStartDate: 'Start Date',
        workEndDate: 'End Date',
        familyName: 'Name',
        familyRelation: 'Relation',
        familyPhone: 'Phone',
        familyWorkUnit: 'Work Unit',
        form: {
          employeeNo: 'Employee No. (auto if empty)',
          employeeNoAuto: 'Leave empty to auto-generate (max+1)',
          name: 'Please enter name',
          org: 'Please select department',
          company: 'Please select company',
          position: 'Please enter position',
          post: 'Please enter job post',
          personStatus: 'Please select person status',
          phone: 'Please enter phone',
          email: 'Please enter email',
          birthday: 'Birthday (e.g. 1990-01-01)',
          nation: 'Please enter nation',
          politicalStatus: 'Please enter political status',
          maritalStatus: 'Please select marital status',
          nativePlace: 'Please enter native place',
          bloodType: 'Please enter blood type',
          idCard: 'Please enter ID card no.',
          householdAddress: 'Please enter household address',
          address: 'Please enter address',
          emergencyContact: 'Please enter emergency contact',
          emergencyPhone: 'Please enter emergency phone',
          bankName: 'Please enter salary bank',
          bankAccount: 'Please enter salary account',
          remark: 'Please enter remark'
        }
      },
      attendance: {
        yes: 'Yes',
        no: 'No',
        status: {
          enabled: 'Enabled',
          disabled: 'Disabled'
        },
        group: {
          title: 'Attendance Groups',
          groupName: 'Group Name',
          groupType: {
            title: 'Group Type',
            fixed: 'Fixed',
            scheduling: 'Scheduling',
            flexible: 'Flexible Hours'
          },
          workTime: 'Work Time',
          workDays: 'Work Days',
          scope: 'Scope',
          scopeDepts: 'Depts×{n}',
          scopeMembers: 'Members×{n}',
          managers: 'Managers',
          status: 'Status',
          remark: 'Remark',
          enable: 'Enable',
          disable: 'Disable',
          confirmEnable: 'Enable this attendance group?',
          confirmDisable: 'Attendance will fall back to the org default rule after disabling. Continue?',
          flexibleNoTime: 'Flexible hours (no clock)',
          schedulingFallback: 'By schedule',
          members: 'Effective Members',
          noMembers: 'No effective members',
          addGroup: 'Add Attendance Group',
          editGroup: 'Edit Attendance Group',
          weekday: {
            1: 'Mon',
            2: 'Tue',
            3: 'Wed',
            4: 'Thu',
            5: 'Fri',
            6: 'Sat',
            7: 'Sun'
          },
          weekdayShort: {
            1: 'Mon',
            2: 'Tue',
            3: 'Wed',
            4: 'Thu',
            5: 'Fri',
            6: 'Sat',
            7: 'Sun'
          },
          form: {
            groupName: 'Please enter group name',
            groupType: 'Please select group type',
            workDays: 'Please select work days',
            orgScope: 'Departments (incl. children)',
            memberScope: 'Designated members',
            managers: 'Managers (press Enter to add)',
            status: 'Please select status',
            remark: 'Please enter remark'
          }
        },
        config: {
          title: 'Org Default Rules',
          org: 'Organization',
          orgName: 'Organization',
          quickStart: 'Quick Start',
          editConfig: 'Edit Org Default Rule',
          recheck: 'Re-check',
          fallbackHint:
            "Rule priority: attendance group > org default rule; employees not in any group follow their own department's rule; if unset, the nearest ancestor organization's rule applies.",
          noRule: 'No rule',
          sourceGroup: 'Group',
          sourceOrg: 'Org rule',
          healthEmployee: 'Employee',
          healthSource: 'Rule Source',
          healthyTotal: '{n} in total',
          healthyGroup: 'Group hit {n}',
          healthyOrg: 'Org rule {n}',
          healthyNoRule: 'No rule {n}',
          form: {
            orgName: 'Please enter organization name',
            org: 'Please select organization'
          }
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
        },
        overtimeRule: {
          title: 'Overtime Rules',
          ruleName: 'Rule Name',
          workday: 'Workday',
          restday: 'Rest Day',
          holiday: 'Statutory Holiday',
          dayPolicy: 'Day Policy',
          approvalRequired: 'Approval',
          punchRequired: 'Punch Requirement',
          compMultiplier: 'Comp Multiplier',
          multiplierTip: 'Multiplier N means 1 hour overtime converts to N hours time-off',
          multiplierValue: 'Comp x{n}',
          needApproval: 'Approval Required',
          noApproval: 'No Approval',
          needPunch: 'Punch Required',
          noPunch: 'No Punch',
          minUnitMinutes: 'Min Unit',
          minUnitHalfHour: '30 min',
          minUnitHour: '1 hour',
          monthlyLimitHours: 'Monthly Limit',
          monthlyLimitTip: '0 means unlimited',
          noLimit: 'Unlimited',
          usedGroups: 'Used Groups',
          boundGroupCount: '{n} group(s)',
          globalFallback: 'All Groups',
          addRule: 'Add Overtime Rule',
          editRule: 'Edit Overtime Rule',
          enable: 'Enable',
          disable: 'Disable',
          status: 'Status',
          remark: 'Remark',
          form: {
            ruleName: 'Please enter rule name',
            status: 'Please select status',
            remark: 'Please enter remark'
          }
        },
        scheduleCycle: {
          title: 'Schedule Cycles',
          cycleName: 'Cycle Name',
          shiftSequence: 'Shift Sequence',
          cycleDays: 'Cycle Days',
          cycleDaysTip: 'Derived from shift sequence length',
          sequenceSummary: '{n} shifts',
          usedGroups: 'Attendance Groups',
          boundGroupCount: '{n} groups',
          noBoundGroup: 'Not Bound',
          addCycle: 'Add Schedule Cycle',
          editCycle: 'Edit Schedule Cycle',
          enable: 'Enable',
          disable: 'Disable',
          status: 'Status',
          remark: 'Remark',
          form: {
            cycleName: 'Please enter cycle name',
            shiftSequence: 'Please select shift sequence (in order)',
            usedGroups: 'Please select attendance groups',
            status: 'Please select status',
            remark: 'Please enter remark'
          }
        }
      },
      employeeRelation: {
        billStatus: {
          title: 'Bill Status',
          unsubmitted: 'Unsubmitted',
          approving: 'Approving',
          approved: 'Approved',
          rejected: 'Rejected',
          revoked: 'Revoked'
        },
        transfer: {
          title: 'Transfer',
          billNo: 'Bill No.',
          employeeNo: 'Employee No.',
          employeeName: 'Employee Name',
          changeType: {
            title: 'Change Type',
            post: 'Position',
            org: 'Department',
            salary: 'Salary',
            company: 'Company',
            other: 'Other'
          },
          originalCompany: 'Original Company',
          originalOrg: 'Original Department',
          originalPosition: 'Original Position Title',
          originalPost: 'Original Post',
          newCompany: 'New Company',
          newOrg: 'New Department',
          newPosition: 'New Position Title',
          newPost: 'New Post',
          newSalary: 'New Salary',
          transferReason: 'Reason',
          effectiveImmediately: 'Effective Immediately',
          effectiveDate: 'Effective Date',
          applied: 'Applied',
          creatorName: 'Applicant',
          createTime: 'Created At',
          yes: 'Yes',
          no: 'No',
          submit: 'Submit',
          approve: 'Approve',
          reject: 'Reject',
          revoke: 'Revoke',
          records: 'Records',
          noRecords: 'No approval records yet',
          confirmSubmit: 'Submit this transfer bill for approval?',
          confirmApprove: 'Approve this bill? Immediately-effective transfers update the employee archive at once.',
          confirmReject: 'Reject this transfer bill?',
          confirmRevoke: 'Revoke this transfer bill?',
          addTransfer: 'Add Transfer',
          editTransfer: 'Edit Transfer',
          sectionEmployee: 'Employee',
          sectionOriginal: 'Original (before)',
          sectionNew: 'New (after)',
          sectionEffect: 'Effectiveness',
          form: {
            billNo: 'Please enter bill no.',
            status: 'Please select bill status',
            employeeNo: 'Please enter employee no.',
            employeeName: 'Please enter employee name',
            changeType: 'Please select change type',
            org: 'Please select org',
            transferReason: 'Please enter reason',
            effectiveDateRequired: 'Effective date is required when not effective immediately'
          }
        },
        entry: {
          title: 'Entry Bills',
          billNo: 'Bill No.',
          billStatus: {
            title: 'Status'
          },
          name: 'Name',
          phone: 'Phone',
          email: 'Email',
          gender: 'Gender',
          birthday: 'Birthday',
          nation: 'Nation',
          politicalStatus: 'Political Status',
          maritalStatus: 'Marital Status',
          cultureLevel: 'Education Level',
          nativePlace: 'Native Place',
          address: 'Current Address',
          householdAddress: 'Household Address',
          emergencyContact: 'Emergency Contact',
          emergencyPhone: 'Emergency Phone',
          bankName: 'Salary Bank',
          bankAccount: 'Bank Account',
          avatar: 'Photo',
          position: 'Position',
          post: 'Job Title',
          probationMonths: 'Probation (months)',
          orgUuid: 'Department',
          companyUuid: 'Company',
          hireDate: 'Hire Date',
          expectFormalDate: 'Expected Formal Date',
          attachmentDesc: 'Attachment Notes',
          remark: 'Remark',
          archived: 'Archived',
          archivedYes: 'Yes',
          archivedNo: 'No',
          creatorName: 'Applicant',
          createTime: 'Created At',
          submit: 'Submit',
          approve: 'Approve',
          reject: 'Reject',
          revoke: 'Revoke',
          records: 'Records',
          noRecords: 'No approval records yet',
          confirmSubmit: 'Submit this entry bill for approval?',
          confirmApprove:
            'Approve this bill? An employee archive will be created automatically (employee no. generated).',
          confirmReject: 'Reject this bill?',
          confirmRevoke: 'Revoke this bill?',
          addEntry: 'Add Entry Bill',
          editEntry: 'Edit Entry Bill',
          tabBasic: 'Bill Info',
          tabWorks: 'Work Experience',
          tabEducations: 'Education',
          tabFamilies: 'Family Members',
          tabAttachments: 'Attachments',
          sectionBasic: 'Basic Info',
          addRow: 'Add Row',
          noRows: 'No rows yet, click the button above to add',
          workCompany: 'Company',
          workPosition: 'Position',
          school: 'School',
          major: 'Major',
          education: 'Degree',
          familyName: 'Name',
          relation: 'Relation',
          familyPhone: 'Phone',
          workUnit: 'Work Unit',
          startDate: 'Start',
          endDate: 'End',
          form: {
            name: 'Please enter name',
            gender: 'Please select gender',
            birthday: 'Please select birthday',
            phone: 'Please enter phone',
            email: 'Please enter email',
            nation: 'Please enter nation',
            politicalStatus: 'Please enter political status',
            maritalStatus: 'Please select marital status',
            cultureLevel: 'Please enter education level',
            nativePlace: 'Please enter native place',
            address: 'Please enter current address',
            householdAddress: 'Please enter household address',
            emergencyContact: 'Please enter emergency contact',
            emergencyPhone: 'Please enter emergency phone',
            bankName: 'Please enter salary bank',
            bankAccount: 'Please enter bank account',
            avatar: 'Please enter photo URL',
            position: 'Please enter position',
            post: 'Please enter job title',
            orgUuid: 'Please enter department',
            companyUuid: 'Please enter company',
            expectFormalDate: 'Leave empty to compute from hire date + probation months',
            attachmentDesc: 'Please describe the attachments to upload',
            remark: 'Please enter remark',
            billNo: 'Please enter bill no.',
            status: 'Please select status'
          }
        }
      },
      salary: {
        subject: {
          title: 'Salary Subjects',
          subjectName: 'Subject Name',
          orgName: 'Linked Company',
          payAccount: 'Pay Account',
          payBank: 'Pay Bank',
          status: 'Status',
          remark: 'Remark',
          addSubject: 'Add Salary Subject',
          editSubject: 'Edit Salary Subject',
          form: {
            subjectName: 'Please enter subject name',
            orgName: 'Please select linked company',
            payAccount: 'Please enter pay account',
            payBank: 'Please enter pay bank',
            status: 'Please select status',
            remark: 'Please enter remark'
          }
        },
        item: {
          title: 'Salary Items',
          itemName: 'Item Name',
          itemCode: 'Item Code',
          itemType: {
            title: 'Item Type',
            fixed: 'Fixed',
            data: 'Data-driven',
            formula: 'Formula'
          },
          itemClass: {
            title: 'Item Class',
            earning: 'Earning',
            deduction: 'Deduction',
            cost: 'Cost'
          },
          valueMode: {
            title: 'Value Mode',
            manual: 'Manual',
            system: 'System',
            formula: 'Formula'
          },
          calcOrder: 'Calc Order',
          orderNum: 'Display Order',
          formula: 'Formula Expression',
          isPreset: 'Preset',
          presetYes: 'Preset',
          presetNo: 'Custom',
          status: 'Status',
          remark: 'Remark',
          addItem: 'Add Salary Item',
          editItem: 'Edit Salary Item',
          form: {
            itemName: 'Please enter item name',
            itemCode: 'Please enter item code (e.g. OTHER_BONUS)',
            itemType: 'Please select item type',
            itemClass: 'Please select item class',
            valueMode: 'Please select value mode',
            formula: 'Please enter formula expression',
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
        noFile: 'No scan file',
        planType: {
          special: 'Special',
          newHire: 'New Hire',
          safety: 'Safety',
          skill: 'Skill'
        },
        published: {
          draft: 'Draft',
          published: 'Published'
        },
        plan: {
          title: 'Training Plan',
          billNo: 'Bill No.',
          billStatusLabel: 'Bill Status',
          planName: 'Plan Name',
          theme: 'Theme',
          planYear: 'Year',
          planType: 'Plan Type',
          startDate: 'Start Date',
          endDate: 'End Date',
          published: 'Publish Status',
          creatorName: 'Applicant',
          addPlan: 'Add Plan',
          editPlan: 'Edit Plan',
          submit: 'Submit',
          approve: 'Approve',
          reject: 'Reject',
          revoke: 'Revoke',
          records: 'Records',
          confirmSubmit: 'Submit this training plan for approval?',
          confirmApprove: 'Approve this plan? It will be published (draft → published) on approval.',
          confirmReject: 'Reject this bill?',
          confirmRevoke: 'Revoke this bill?',
          form: {
            planName: 'Please enter plan name',
            theme: 'Please enter theme',
            planYear: 'Please enter year (e.g. 2026)',
            planType: 'Please select plan type',
            status: 'Please select bill status',
            published: 'Please select publish status',
            endDateBeforeStart: 'End date cannot be earlier than start date'
          }
        },
        session: {
          title: 'Training Session',
          sessionName: 'Session Name',
          planName: 'Training Plan',
          lecturer: 'Lecturer',
          remark: 'Remark',
          addSession: 'Add Session',
          editSession: 'Edit Session',
          participants: 'Participants',
          finish: 'Finish',
          confirmFinish: 'Finish this session? This cannot be undone.',
          participantsTip: 'Saving replaces the full participant list (selection is the final roster).',
          form: {
            sessionName: 'Please enter session name',
            planUuid: 'Please select a published training plan',
            lecturer: 'Please enter lecturer',
            location: 'Please enter location',
            status: 'Please select status',
            participantsPlaceholder: 'Search and select users (account/name)',
            endTimeBeforeStart: 'End time must be later than start time'
          }
        },
        certTemplate: {
          title: 'Certificate Template',
          templateName: 'Template Name',
          certType: 'Certificate Type',
          sourceLabel: 'Source',
          noPrefix: 'No. Prefix',
          statusLabel: 'Status',
          remark: 'Remark',
          addTemplate: 'Add Template',
          editTemplate: 'Edit Template',
          enable: 'Enable',
          disable: 'Disable',
          confirmEnable: 'Enable this template? It can issue certificates once enabled.',
          confirmDisable: 'Disable this template? Disabled templates cannot issue certificates.',
          source: {
            system: 'System',
            org: 'Enterprise'
          },
          status: {
            enabled: 'Enabled',
            disabled: 'Disabled'
          },
          form: {
            templateName: 'Please enter template name',
            certType: 'Please enter certificate type',
            source: 'Please select source',
            noPrefix: 'Please enter number prefix (e.g. JN/AQ/HF; cert no = prefix-year-seq)'
          }
        },
        certificate: {
          title: 'Certificate Issue',
          student: 'Student',
          template: 'Template',
          sourceLabel: 'Source',
          source: {
            session: 'Session Completion',
            import: 'External Import'
          },
          session: 'Session',
          batchIssue: 'Batch Issue',
          batchIssueDone: 'Issued',
          batchIssueSkipped: 'Skipped (already certified)',
          batchIssueTip:
            'Generates certificates for the finished-session participants (cert no = template prefix-year-seq); students already holding a valid certificate of the template are skipped (idempotent).',
          importExternal: 'Import External Certificate',
          issue: 'Issue',
          revoke: 'Revoke',
          upload: 'Upload Scan',
          chooseFile: 'Choose File',
          confirmIssue: 'Issue this certificate? It will be visible in My Training.',
          confirmRevoke: 'Revoke this certificate? Revocation is terminal.',
          form: {
            session: 'Please select a session (finished)',
            template: 'Please select a certificate template (enabled)',
            student: 'Search and select a student (account/name)',
            certName: 'Please enter certificate name',
            certType: 'Please enter certificate type',
            certNo: 'Please enter external cert no (unique)',
            status: 'Please select status',
            source: 'Please select source'
          }
        }
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
